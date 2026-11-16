package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/global_record_original_v1.json.gz
var globalRecordNativeOriginal []byte

type globalRecordNativeCase struct {
	Input struct {
		Name             string
		Index            uint32
		Layout           int
		SourceMode       string `json:"source_mode"`
		SourceAlignment  int    `json:"source_alignment"`
		SourceHex        string `json:"source_hex"`
		InitialBufferHex string `json:"initial_buffer_hex"`
		StateWord        uint32 `json:"state_word"`
	}
	Returns      [5]*uint32
	BufferAfter  string `json:"buffer_after"`
	OutputStores []struct {
		Slot    int
		Pointer uint32
	} `json:"output_stores"`
}

func globalRecordNativeCases(t *testing.T) []globalRecordNativeCase {
	t.Helper()
	var cases []globalRecordNativeCase
	npcVoiceOriginalDecode(t, globalRecordNativeOriginal, "d5e527a06bf8e0126b4258d0563010675120a8f6235c21e3ade6bac60cfc8fef", "52aae2490398d507e2d1b79f1953756617ee29d672f15d54e147243f82c876f1", 1088, &cases)
	return cases
}

func TestGlobalRecordNativeLayout(t *testing.T) {
	word := unsafe.Sizeof(uintptr(0))
	want := [13]uintptr{word, 4, 4, 1, 4, word, word, word, word, word, unsafe.Alignof(uintptr(0)), 24, 10}
	if got := globalRecordNativeLayoutC(); got != want {
		t.Fatal("global record scalar/pointer/function ABI", got, want)
	}
	pointers := globalRecordNativeOwnerPointersC()
	wantOwners := [4]uintptr{
		uintptr(memmap.PtrOff(0x5D4594, 0x650000-0x5D4594)),
		uintptr(memmap.PtrOff(0x5D4594, 531660)),
		uintptr(memmap.PtrOff(0x5D4594, 534756)),
		uintptr(memmap.PtrOff(0x587000, 60072)),
	}
	for i, value := range pointers {
		if i < len(wantOwners) && value != wantOwners[i] {
			t.Fatal("C/Go shared global owner identity", i, value, wantOwners[i])
		}
		if value == 0 || word == 8 && value <= math.MaxUint32 {
			t.Fatal("full native global pointer required", i, value)
		}
	}
}

func globalRecordNativeBytes(t *testing.T, text string) []byte {
	t.Helper()
	out, err := hex.DecodeString(text)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func globalRecordNativeRun(t *testing.T, routes []int, subtests bool, extra string) {
	t.Helper()
	cases := globalRecordNativeCases(t)
	global := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 0x650000-0x5D4594)), 65536)
	stateOwner := memmap.Slice(0x587000, 60032)[:64]
	savedGlobal := append([]byte(nil), global...)
	savedStateOwner := append([]byte(nil), stateOwner...)
	savedIndex := globalRecordNativeSelectedC(false, 0)
	t.Cleanup(func() {
		copy(global, savedGlobal)
		copy(stateOwner, savedStateOwner)
		globalRecordNativeSelectedC(true, savedIndex)
	})
	word := int(unsafe.Sizeof(uintptr(0)))
	for _, route := range routes {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if word == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign output/source owner above4GiB required")
			}
			for _, c := range cases {
				check := func(t *testing.T) {
					initial := globalRecordNativeBytes(t, c.Input.InitialBufferHex)
					payload := append(globalRecordNativeBytes(t, c.Input.SourceHex), 0)
					after := globalRecordNativeBytes(t, c.BufferAfter)
					if len(initial) != 52 || len(after) != 52 || len(payload) > 52 || c.Returns[0] == nil || *c.Returns[0] != 0x656E78 || c.Returns[1] != nil || c.Input.Layout < 0 || c.Input.Layout > 6 {
						t.Fatal("invalid original global record domain", c.Input.Name)
					}
					for j := range raw {
						raw[j] = byte(j*37 + 29)
						global[j] = byte(j*53 + 17)
					}
					for j := range stateOwner {
						stateOwner[j] = byte(j*31 + 7)
					}
					copy(global[0x6E78:], initial)
					sourceAt := 0x1000 + placement + c.Input.SourceAlignment
					copy(raw[sourceAt:], payload)
					var source unsafe.Pointer
					switch c.Input.SourceMode {
					case "null":
					case "self":
						source = unsafe.Pointer(&global[0x6E78])
					case "foreign":
						source = unsafe.Pointer(&raw[sourceAt])
					default:
						t.Fatal("invalid original source role")
					}
					layouts := [7][2]int{{0, 16}, {0, 0}, {0, word}, {16, 0}, {-1, -1}, {-1, 0}, {0, -1}}
					var slots [2]unsafe.Pointer
					for i, offset := range layouts[c.Input.Layout] {
						if offset >= 0 {
							slots[i] = unsafe.Pointer(&raw[0x2000+placement+offset])
						}
					}
					globalRecordNativeSelectedC(true, c.Input.Index)
					globalRecordNativeStateC(true, c.Input.StateWord)
					wantGlobal := append([]byte(nil), global...)
					wantRaw := append([]byte(nil), raw...)
					wantState := append([]byte(nil), stateOwner...)
					project := func(pointer uint32) uintptr {
						if pointer < 0x650000 || pointer >= 0x660000 {
							t.Fatal("original pointer outside declared global owner", pointer)
						}
						return uintptr(unsafe.Pointer(&global[pointer-0x650000]))
					}
					verify := func() {
						if !bytes.Equal(raw, wantRaw) || !bytes.Equal(global, wantGlobal) || !bytes.Equal(stateOwner, wantState) || globalRecordNativeSelectedC(false, 0) != c.Input.Index || globalRecordNativeStateC(false, 0) != c.Input.StateWord {
							t.Fatal("whole64KiB global/foreign owner, scalar guards or selected-index shadow changed", c.Input.Name)
						}
					}
					for entry, original := range c.Returns {
						if entry == 1 {
							copy(wantGlobal[0x6E78:], after)
						}
						if entry == 2 {
							for _, store := range c.OutputStores {
								if store.Slot < 0 || store.Slot > 1 || slots[store.Slot] == nil {
									t.Fatal("invalid original pointer store", c.Input.Name)
								}
								at := 0x2000 + placement + layouts[c.Input.Layout][store.Slot]
								value := project(store.Pointer)
								if word == 8 {
									binary.LittleEndian.PutUint64(wantRaw[at:], uint64(value))
								} else {
									binary.LittleEndian.PutUint32(wantRaw[at:], uint32(value))
								}
							}
						}
						got := globalRecordNativeCall(route, entry, source, slots[0], slots[1])
						if original != nil {
							want := uintptr(*original)
							if entry == 0 {
								want = project(*original)
							}
							if got != want {
								t.Fatalf("%s route%d place%d entry%d: got%#x want%#x", c.Input.Name, route, placement, entry, got, want)
							}
						}
						verify()
					}
					if extra == "wrapper" {
						end := bytes.IndexByte(after, 0)
						if end < 0 || Sub_41FA40() != string(after[:end]) {
							t.Fatal("existing Go global-string consumer", c.Input.Name)
						}
						verify()
					}
					if extra == "parent" {
						at := 0x4000 + placement
						clear(wantRaw[at : at+108])
						got := globalRecordNativeParentC(unsafe.Pointer(&raw[at]))
						want := [4]uintptr{uintptr(*c.Returns[2]), 0, 0, 0}
						if want[0] != 0 {
							if len(c.OutputStores) != 2 {
								t.Fatal("two original pointer identities required")
							}
							want[1], want[2], want[3] = project(c.OutputStores[0].Pointer), project(c.OutputStores[1].Pointer), 1
						}
						if got != want {
							t.Fatal("actual bounded C parent pointer slots/null owner service", c.Input.Name, got, want)
						}
						verify()
					}
				}
				if subtests {
					t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), check)
				} else {
					check(t)
				}
			}
		}
	}
}

func TestGlobalRecordNativeOriginal(t *testing.T) {
	globalRecordNativeRun(t, []int{0, 1, 2}, true, "")
}

func TestGlobalRecordNativeWrapper(t *testing.T) {
	globalRecordNativeRun(t, []int{2}, false, "wrapper")
}

func TestGlobalRecordNativeBoundedCParent(t *testing.T) {
	globalRecordNativeRun(t, []int{2}, false, "parent")
}
