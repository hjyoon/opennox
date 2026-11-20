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

//go:embed testdata/global_state_original_v1.json.gz
var globalStateNativeOriginal []byte

type globalStateNativeCase struct {
	Input struct {
		Index       int
		Entry, Name string
		Nonnull     bool
		Word        uint32
		ActiveWord  uint32 `json:"active_word"`
		Flags       uint32
		Level       byte
	}
	ReturnWord  *uint32  `json:"return_u32"`
	FinalGlobal uint32   `json:"final_global_u32"`
	Writes      []uint32 `json:"global_write_values"`
}

func TestGlobalStateNativeLayout(t *testing.T) {
	var player server.Player
	pointer := unsafe.Sizeof(uintptr(0))
	cellSize := uintptr(148)
	if pointer == 8 {
		cellSize = 152
	}
	want := [17]uintptr{pointer, unsafe.Sizeof(player), unsafe.Offsetof(player.Active), 1, 3, unsafe.Offsetof(player.Field3680), 4, unsafe.Offsetof(player.Level), 1, cellSize, pointer, 4, pointer + 16, pointer, pointer + 4, 4, 4}
	if got := globalStateNativeLayoutC(); got != want {
		t.Fatal("native global/player/inventory/window layouts", got, want)
	}
	if unsafe.Offsetof(player.Active)%4 != 0 || unsafe.Offsetof(player.Field2096Buf) != unsafe.Offsetof(player.Active)+4 || unsafe.Sizeof(player) > 0x2000 {
		t.Fatal("original active-word span or native owner stride")
	}
}

func globalStateNativePointerSlot(index int) bool {
	return index < 4 || index == 8 || index == 9 || index == 12
}

func globalStateNativeSlotBytes(index int) []byte {
	size := 4
	if globalStateNativePointerSlot(index) {
		size = int(unsafe.Sizeof(uintptr(0)))
	}
	return unsafe.Slice((*byte)(globalStateNativeSlot(index)), size)
}

func TestGlobalStateNativeOriginal(t *testing.T) {
	var cases []globalStateNativeCase
	npcVoiceOriginalDecode(t, globalStateNativeOriginal, "451b71b00abd252308077ea1ff1ec939a8a90f2ecd7d7460894ef248e540e4d8", "9b909bb0d714d15af0976eeaa56694b2d82e56c19ef0237c224983d631679d0f", 2926, &cases)
	var saved [17][]byte
	for i := range saved {
		saved[i] = append([]byte(nil), globalStateNativeSlotBytes(i)...)
	}
	windows := globalStateNativeWindows()
	savedWindows := append([]byte(nil), windows...)
	defer func() {
		for i := range saved {
			copy(globalStateNativeSlotBytes(i), saved[i])
		}
		copy(windows, savedWindows)
	}()
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign owner above4GiB required")
			}
			player := (*server.Player)(unsafe.Pointer(&raw[0x1000+placement]))
			cell := unsafe.Pointer(&raw[0x4000+placement])
			before := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i >= 17 || (c.ReturnWord == nil) != (i >= 10 && i <= 14) {
						t.Fatal("invalid sealed original entry")
					}
					for j := range raw {
						raw[j] = byte(j*47 + 23)
					}
					// C owns these slots and all seeded pointer bytes. Never write a Go
					// pointer over the arbitrary previous canary in a foreign record.
					for j := range saved {
						if globalStateNativePointerSlot(j) {
							value := unsafe.Pointer(&raw[0x6000+j*64+placement])
							if (j < 3 && !c.Input.Nonnull) || (j >= 3 && c.Input.Word == 0) {
								value = nil
							}
							globalStateNativeStorePointer(globalStateNativeSlot(j), value)
						} else {
							*globalStateNativeWord(j) = 0xabcdef01 + uint32(j)
						}
					}
					var playerPtr, cellPtr unsafe.Pointer
					if c.Input.Nonnull {
						playerPtr = unsafe.Pointer(player)
						cellPtr = cell
					}
					globalStateNativeStorePointer(globalStateNativeSlot(0), playerPtr)
					globalStateNativeStorePointer(globalStateNativeSlot(2), cellPtr)
					*(*uint32)(unsafe.Pointer(&player.Active)) = c.Input.ActiveWord
					player.Field3680 = c.Input.Flags
					player.Level = c.Input.Level
					globalStateNativeSeedCell(cell, c.Input.Word)
					if !globalStateNativePointerSlot(i) {
						*globalStateNativeWord(i) = c.Input.Word
					}
					copy(before, raw)
					var snapshots [17][]byte
					for j := range snapshots {
						snapshots[j] = append([]byte(nil), globalStateNativeSlotBytes(j)...)
					}
					windowBefore := append([]byte(nil), windows...)
					result := uint32(globalStateNativeCall(route, i))
					if c.ReturnWord != nil && result != *c.ReturnWord {
						t.Fatalf("original int32 return %#x want %#x", result, *c.ReturnWord)
					}
					if i == 12 {
						if c.FinalGlobal != 0 || globalStateNativePointer(i) != nil {
							t.Fatal("native waypoint slot not cleared")
						}
					} else if i >= 10 && i <= 14 {
						if got := *globalStateNativeWord(i); got != c.FinalGlobal {
							t.Fatalf("original final global %#x want %#x", got, c.FinalGlobal)
						}
					}
					for j := range snapshots {
						if i >= 10 && i <= 14 && globalStateNativeSlot(j) == globalStateNativeSlot(i) {
							continue
						}
						if !bytes.Equal(globalStateNativeSlotBytes(j), snapshots[j]) {
							t.Fatalf("unrelated native global%d changed", j)
						}
					}
					if !bytes.Equal(raw, before) || !bytes.Equal(windows, windowBefore) {
						t.Fatal("whole foreign owners or seven windows changed")
					}
				})
			}
		}
	}
}
