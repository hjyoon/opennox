package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/flag_predicates_original_v1.json.gz
var flagPredicateOriginal []byte

type flagPredicateOriginalCase struct {
	Input struct {
		Entry    string
		Name     string
		Present  bool
		Value    uint32
		Mask     byte
		Seed     byte
		OpaqueID uint32 `json:"opaque_id"`
	}
	OriginalFault bool    `json:"original_fault"`
	Return        *uint32 `json:"return_u32"`
	Compatibility *uint32 `json:"incoming_C_nil_compatibility_return"`
}

func TestFlagPredicateNativeLayout(t *testing.T) {
	var win gui.Window
	var wp server.Waypoint
	word := unsafe.Sizeof(uintptr(0))
	want := [11]uintptr{word, unsafe.Sizeof(win), unsafe.Alignof(win), unsafe.Offsetof(win.Flags), unsafe.Sizeof(win.Flags), unsafe.Sizeof(wp), unsafe.Alignof(wp), unsafe.Offsetof(wp.Flags2), unsafe.Sizeof(wp.Flags2), word, word}
	if got := flagPredicateNativeLayoutC(); got != want || got[3] != 4 {
		t.Fatal("native C/Go owners, scalar offsets and public function types", got, want)
	}
}

func TestFlagPredicateNativeOriginal(t *testing.T) {
	var cases []flagPredicateOriginalCase
	npcVoiceOriginalDecode(t, flagPredicateOriginal,
		"d9a53e1591b80fb4206a49927b70a9e6cda838391af781fa5db7a82a27b6205f",
		"c29abc4b9a67932c1f310b42942742febbe2654f1c66a9f48537371b38894963", 66118, &cases)
	const ownerBytes, guard = 65536, 32
	patterns := make([][]byte, 256)
	for seed := range patterns {
		patterns[seed] = make([]byte, ownerBytes)
		for i := range patterns[seed] {
			patterns[seed][i] = byte(seed + 17*i + 3*(i>>8))
		}
	}
	for route := 0; route < 3; route++ {
		for _, place := range []int{0x100, 0x2000, 0x8000} {
			t.Run(fmt.Sprintf("route%d/place%x", route, place), func(t *testing.T) {
				buffer, free := alloc.MakeHigh([]byte{}, ownerBytes+2*guard)
				defer free()
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&buffer[0])) <= math.MaxUint32 {
					t.Fatal("flag predicate owner must exceed 4GiB")
				}
				for i := range buffer {
					buffer[i] = 0xd7
				}
				expected := make([]byte, len(buffer))
				owner := unsafe.Pointer(&buffer[guard+place])
				var windows, pairs, compatibility int
				for _, c := range cases {
					copy(buffer[guard:guard+ownerBytes], patterns[c.Input.Seed])
					kind := 0
					if c.Input.Entry == "waypoint" {
						kind = 1
					} else if c.Input.Entry != "window" {
						t.Fatal("undeclared entry", c.Input.Entry)
					}
					flagPredicateNativeSeedC(kind, owner, c.Input.Value, c.Input.OpaqueID)
					copy(expected, buffer)
					argument := owner
					if !c.Input.Present {
						argument = nil
					}
					var wanted uint32
					if c.OriginalFault {
						// Existing C adds a null guard to the original 579EE0.
						// This checks compatibility, not original null-fault parity.
						if kind != 1 || c.Input.Present || c.Return != nil || c.Compatibility == nil || *c.Compatibility != 0 {
							t.Fatal("undeclared nil compatibility context", c.Input.Name)
						}
						compatibility++
						wanted = *c.Compatibility
					} else {
						if c.Return == nil || c.Compatibility != nil {
							t.Fatal("missing original return", c.Input.Name)
						}
						wanted = *c.Return
						if kind == 0 {
							windows++
						} else {
							pairs++
						}
					}
					if got := flagPredicateNativeCall(route, kind, argument, c.Input.Mask); got != wanted {
						t.Fatal("complete return differs", c.Input.Name, got, wanted)
					}
					if !bytes.Equal(buffer, expected) {
						t.Fatal("unexpected write in whole foreign owner or guards", c.Input.Name)
					}
				}
				if windows != 326 || pairs != 65536 || compatibility != 256 {
					t.Fatal("original and separate compatibility domains", windows, pairs, compatibility)
				}
			})
		}
	}
}
