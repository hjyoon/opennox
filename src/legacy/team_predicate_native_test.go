package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/team_predicates_original_v1.json.gz
var teamPredicateNativeOriginal []byte

type teamPredicateNativeCase struct {
	Input struct {
		Entry       string
		Name        string
		FirstRole   int `json:"first_role"`
		SecondRole  int `json:"second_role"`
		Records     [2][8]byte
		PatternSeed int `json:"pattern_seed"`
	}
	Return uint32 `json:"return_u32"`
}

func TestTeamPredicateNativeLayout(t *testing.T) {
	var value server.ObjectTeam
	word := unsafe.Sizeof(uintptr(0))
	want := [11]uintptr{word, unsafe.Sizeof(value), unsafe.Alignof(value), unsafe.Offsetof(value.Field0), unsafe.Offsetof(value.ID), unsafe.Offsetof(value.ID) + 1, unsafe.Sizeof(value.Field0), unsafe.Sizeof(value.ID), 3, word, word}
	got := teamPredicateNativeLayoutC()
	if got != want || got[1] != 8 || got[4] != 4 {
		t.Fatal("native C/Go record and both public function types", got, want)
	}
}

func TestTeamPredicateNativeOriginal(t *testing.T) {
	var cases []teamPredicateNativeCase
	npcVoiceOriginalDecode(t, teamPredicateNativeOriginal,
		"bcd8ae0b12802808363921c2ffb4feba02261615a7b50af7026847708c6102d8",
		"436cd0bc8aaf46c9355889c648ea308395f27c1824633baa0b31740c415b73a2", 4078, &cases)
	const ownerBytes, guard = 65536, 32
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			var owners, expected [2][]byte
			var records [2]*server.ObjectTeam
			for owner := range owners {
				buffer, free := alloc.MakeHigh([]byte{}, ownerBytes+2*guard)
				t.Cleanup(free)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&buffer[0])) <= math.MaxUint32 {
					t.Fatal("team predicate owner must exceed4GiB")
				}
				owners[owner] = buffer
				expected[owner] = make([]byte, len(buffer))
				records[owner] = (*server.ObjectTeam)(unsafe.Pointer(&buffer[guard+4096+placement]))
			}
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					kind := 0
					if c.Input.Entry == "same" {
						kind = 1
					} else if c.Input.Entry != "has" {
						t.Fatal("undeclared original entry", c.Input.Entry)
					}
					for owner, buffer := range owners {
						for i := range buffer {
							buffer[i] = byte(0xd7 + owner)
						}
						for i := 0; i < ownerBytes; i++ {
							buffer[guard+i] = byte(i*37 + c.Input.PatternSeed + owner*97)
						}
						copy(buffer[guard+4096+placement:], c.Input.Records[owner][:])
						copy(expected[owner], buffer)
					}
					role := func(value int) *server.ObjectTeam {
						if value == 0 {
							return nil
						}
						if value < 1 || value > 2 {
							t.Fatal("undeclared pointer role", value)
						}
						return records[value-1]
					}
					got := teamPredicateNativeCall(route, kind, role(c.Input.FirstRole), role(c.Input.SecondRole))
					if uint32(got) != c.Return {
						t.Fatal("original complete return differs", got, c.Return)
					}
					for owner, buffer := range owners {
						if !bytes.Equal(buffer, expected[owner]) {
							t.Fatal("unexpected write in whole foreign owner, opaque record or guards", owner)
						}
					}
				})
			}
		}
	}
}
