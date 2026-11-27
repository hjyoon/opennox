package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/gui_bar_state_original_v1.json.gz
var guiBarStateNativeOriginal []byte

var guiBarStateNativeWordNames = [...]string{
	"base_color2", "base_color1", "alternate_color1", "alternate_color2",
	"health", "health_max", "color1", "color2", "mana", "mana_max",
	"refresh", "mode", "player_mana_max", "player_health_max",
}

type guiBarStateNativeCase struct {
	Input struct {
		Index       int
		Entry, Name string
		A, B        uint32
		Nonnull     bool
		Words       map[string]uint32
	}
	Return uint32            `json:"return_u32"`
	Final  map[string]uint32 `json:"final_words"`
}

func TestGUIBarStateNativeLayout(t *testing.T) {
	var player server.Player
	pointer := unsafe.Sizeof(uintptr(0))
	info, ok := reflect.TypeOf(player).FieldByName("info")
	if !ok {
		t.Fatal("Player info field missing")
	}
	infoType := reflect.TypeOf(*player.Info())
	mana, manaOK := infoType.FieldByName("field2243")
	health, healthOK := infoType.FieldByName("field2247")
	if !manaOK || !healthOK {
		t.Fatal("PlayerInfo maximum fields missing")
	}
	want := [17]uintptr{pointer, unsafe.Sizeof(player), info.Offset, info.Type.Size(), info.Offset + mana.Offset, info.Offset + health.Offset, 4, pointer + 16, pointer, pointer + 4, pointer + 8, pointer + 12, 7 * (pointer + 16), 4, pointer, health.Type.Size(), mana.Type.Size()}
	if got := guiBarStateNativeLayoutC(); got != want {
		t.Fatal("native GUI/Player field projection", got, want)
	}
	if unsafe.Sizeof(player) > 0x2000 || mana.Type.Size() != 4 || health.Type.Size() != 4 {
		t.Fatal("native Player owner span or exact scalar widths")
	}
}

func TestGUIBarStateNativeOriginal(t *testing.T) {
	var cases []guiBarStateNativeCase
	npcVoiceOriginalDecode(t, guiBarStateNativeOriginal, "588c3f8388c387f9953ef9ace6db49dc0f999c3ca9113a1fd6de76162d4d74f6", "e09c86ede6de9d7c6eddcf1328b4a38878f3bd8c93ad3900df93c6b878c0a376", 1616, &cases)
	var saved [12]uint32
	for i := range saved {
		saved[i] = *guiBarStateNativeWord(i)
	}
	pointerSlot := globalStateNativeSlotBytes(0)
	savedPointer := append([]byte(nil), pointerSlot...)
	windows := globalStateNativeWindows()
	savedWindows := append([]byte(nil), windows...)
	defer func() {
		for i, value := range saved {
			*guiBarStateNativeWord(i) = value
		}
		copy(pointerSlot, savedPointer)
		copy(windows, savedWindows)
	}()
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("actual foreign Player owner above4GiB required")
			}
			player := unsafe.Pointer(&raw[0x1000+placement])
			var slots [14]*uint32
			for i := 0; i < 12; i++ {
				slots[i] = guiBarStateNativeWord(i)
			}
			for i := 12; i < 14; i++ {
				slots[i] = guiBarStateNativePlayerWordC(player, i)
			}
			before := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					i := c.Input.Index
					if i < 0 || i >= 6 || len(c.Input.Words) != 14 || len(c.Final) != 14 {
						t.Fatal("invalid sealed original case")
					}
					for j := range raw {
						raw[j] = byte(j*47 + 23)
					}
					for j := range windows {
						windows[j] = byte(j*29 + 11)
					}
					for j, name := range guiBarStateNativeWordNames {
						*slots[j] = c.Input.Words[name]
					}
					value := player
					if !c.Input.Nonnull {
						value = nil
					}
					storeForeignPointer(globalStateNativeSlot(0), value)
					pointerBefore := append([]byte(nil), pointerSlot...)
					copy(before, raw)
					windowBefore := append([]byte(nil), windows...)
					got := uint32(guiBarStateNativeCall(route, i, int32(c.Input.A), int32(c.Input.B)))
					if got != c.Return {
						t.Fatalf("original int32 return %#x want %#x", got, c.Return)
					}
					for j, name := range guiBarStateNativeWordNames {
						want := c.Final[name]
						if *slots[j] != want {
							t.Fatalf("original final %s %#x want %#x", name, *slots[j], want)
						}
						// Permit only words actually written by this original root,
						// projected into the existing native Player/window owners.
						allowed := (i == 0 && (j == 6 || j == 7 || j == 11)) ||
							(i == 1 && (j == 4 || j == 5 || j == 10 || (j == 13 && c.Input.Nonnull))) ||
							(i == 2 && j == 4) ||
							(i == 3 && (j == 8 || j == 9 || j == 10 || (j == 12 && c.Input.Nonnull))) ||
							(i == 4 && j == 8)
						if !allowed {
							continue
						}
						at := uintptr(unsafe.Pointer(slots[j]))
						for _, owner := range []struct {
							base uintptr
							data []byte
						}{
							{uintptr(unsafe.Pointer(&raw[0])), before},
							{uintptr(unsafe.Pointer(&windows[0])), windowBefore},
						} {
							if at >= owner.base && at+4 <= owner.base+uintptr(len(owner.data)) {
								binary.LittleEndian.PutUint32(owner.data[at-owner.base:at-owner.base+4], want)
							}
						}
					}
					if !bytes.Equal(raw, before) || !bytes.Equal(windows, windowBefore) || !bytes.Equal(pointerSlot, pointerBefore) {
						t.Fatal("whole foreign Player owner, seven windows or shared Player pointer changed outside original scalar writes")
					}
				})
			}
		}
	}
}
