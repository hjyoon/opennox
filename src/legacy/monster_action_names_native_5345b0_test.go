package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/monster_action_names_original_5345b0_534670.json.gz
var monsterActionNamesOriginal5345B0 []byte

type monsterActionNameOriginalCase5345B0 struct {
	Kind     string          `json:"kind"`
	Value    uint32          `json:"value"`
	InputHex string          `json:"input_hex"`
	Result   json.RawMessage `json:"result"`
}

func monsterActionNameOriginalCases5345B0(t *testing.T) []monsterActionNameOriginalCase5345B0 {
	t.Helper()
	if got := fmt.Sprintf("%x", sha256.Sum256(monsterActionNamesOriginal5345B0)); got != "7236449e4178f8361cbf1a4c2be7aa9e35ece7544af6325880775ece45ad2929" {
		t.Fatal("sealed original compressed action names changed", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(monsterActionNamesOriginal5345B0))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "bcc58b1bfc5ef069886cba104017de5cc69cd835b00631f5e0ca58dffa96da2e" {
		t.Fatal("sealed original action names changed", got)
	}
	var cases []monsterActionNameOriginalCase5345B0
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, c := range cases {
		counts[c.Kind]++
	}
	if len(cases) != 1014 || len(counts) != 4 {
		t.Fatal("original action name case count", len(cases), counts)
	}
	for kind, want := range map[string]int{"selected_name": 101, "name": 101, "action_by_name": 406, "by_name": 406} {
		if counts[kind] != want {
			t.Fatal("original action name operation count", kind, counts[kind], want)
		}
	}
	return cases
}

func TestMonsterActionNamesNativeOriginalCABI(t *testing.T) {
	cases := monsterActionNameOriginalCases5345B0(t)
	t.Logf("unchanged original action-name cases=%d, instruction bytes=260, declared services=0", len(cases))
	for _, placement := range []int{0, 8, 24} {
		t.Run(fmt.Sprintf("placement-%d", placement), func(t *testing.T) {
			owner, free := alloc.MakeHigh([]byte{}, 4096)
			defer free()
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&owner[0])) <= math.MaxUint32 {
				t.Fatal("native C name input must have a foreign high address")
			}
			start := 32 + placement
			for i, c := range cases {
				t.Run(fmt.Sprintf("%04d-%s", i, c.Kind), func(t *testing.T) {
					for i := range owner {
						owner[i] = 0xa5
					}
					switch c.Kind {
					case "selected_name", "name":
						var want *string
						if err := json.Unmarshal(c.Result, &want); err != nil {
							t.Fatal(err)
						}
						got := monsterActionNamesGetC(c.Kind == "selected_name", c.Value)
						if want == nil {
							if got != nil {
								t.Fatalf("C name outside the table=%p, want NULL", got)
							}
						} else {
							if got == nil || alloc.GoString((*byte)(got)) != *want {
								t.Fatalf("C name for %#x=%p, want %q", c.Value, got, *want)
							}
							// The original getters return the same stable table entry
							// for equal names, including a non-NULL empty slot39.
							expected := unsafe.Pointer(alloc.InternCString(*want))
							if got != expected || monsterActionNamesGetC(c.Kind == "selected_name", c.Value) != got {
								t.Fatal("C getter lost the complete stable string pointer")
							}
						}
						for i, b := range owner {
							if b != 0xa5 {
								t.Fatalf("C getter changed an independent owner at %d", i)
							}
						}
					case "action_by_name", "by_name":
						input, err := hex.DecodeString(c.InputHex)
						if err != nil || start+len(input)+1 >= len(owner) {
							t.Fatal("invalid original input", err)
						}
						copy(owner[start:], input)
						owner[start+len(input)] = 0
						before := append([]byte(nil), owner...)
						var want int32
						if err := json.Unmarshal(c.Result, &want); err != nil {
							t.Fatal(err)
						}
						pointer := unsafe.Pointer(&owner[start])
						if got := monsterActionNamesLookupC(c.Kind == "action_by_name", pointer); got != want {
							t.Fatalf("actual const C lookup %s=%d want=%d", c.Kind, got, want)
						}
						if !bytes.Equal(owner, before) {
							t.Fatal("read-only C lookup changed input, suffix or canaries")
						}
					default:
						t.Fatal("unknown original operation", c.Kind)
					}
				})
			}
		})
	}
}
