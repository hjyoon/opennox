package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/npc_equipment_original_52ba70_v1.json.gz
var npcEquipmentNativeOriginal []byte

type npcEquipmentNativeCase struct {
	Input struct {
		Name   string
		Owner  int
		Chain  []int
		Fields [7][3]uint32
	}
	Before, After [7][3]uint32
	Writes        []struct {
		Slot          int
		Before, After uint32
	}
	Return uint32 `json:"return_u32"`
}

func TestNPCEquipmentNativeLayout(t *testing.T) {
	var obj server.Object
	want := [14]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(obj), unsafe.Offsetof(obj.ObjClass), unsafe.Offsetof(obj.ObjSubClass), unsafe.Offsetof(obj.ObjFlags), unsafe.Offsetof(obj.InvNextItem), unsafe.Offsetof(obj.InvFirstItem), unsafe.Sizeof(obj.ObjClass), unsafe.Sizeof(obj.ObjSubClass), unsafe.Sizeof(obj.ObjFlags), unsafe.Sizeof(obj.InvNextItem), unsafe.Sizeof(obj.InvFirstItem), unsafe.Alignof(obj), unsafe.Sizeof(uintptr(0))}
	got := npcEquipmentNativeLayoutC()
	if got[1]+512 > 0x1000 || want[1]+512 > 0x1000 || got[1] < got[6]+got[11] {
		t.Fatal("native record extent", got, want)
	}
	for i := range want {
		if i != 1 && got[i] != want[i] {
			t.Fatal("native C/Go class/subclass/flags/links layout", got, want)
		}
	}
}

// Compare all observed original final fields and returns on actual C, typed C
// callback and direct Go paths. Original ordered writes define expected bytes;
// the native calls additionally preserve every other byte in the foreign owner.
func TestNPCEquipmentNativeOriginal(t *testing.T) {
	var cases []npcEquipmentNativeCase
	npcVoiceOriginalDecode(t, npcEquipmentNativeOriginal, "470e5ce4f5c69b195289b72b8640a5006c8bcff7bed60335e4e4a2ca3c775182", "14eb472f955e3efbe2d88085d575c5c5b1c0bb6240603f23d5c91d417ed3b149", 1233, &cases)
	offsets := [3]uintptr{unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.Object{}.ObjSubClass), unsafe.Offsetof(server.Object{}.ObjFlags)}
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign inventory owner must exceed4GiB")
			}
			var objects [7]*server.Object
			for i := range objects {
				objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
			}
			expected := make([]byte, len(raw))
			words := func() (out [7][3]uint32) {
				for i, obj := range objects {
					out[i] = [3]uint32{uint32(obj.ObjClass), uint32(obj.ObjSubClass), uint32(obj.ObjFlags)}
				}
				return
			}
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					for i, obj := range objects {
						base := 0x1000 + i*0x1000 + placement
						for j, value := range c.Input.Fields[i] {
							binary.LittleEndian.PutUint32(raw[base+int(offsets[j]):], value)
						}
						obj.InvFirstItem, obj.InvNextItem = nil, nil
					}
					if len(c.Input.Chain) > 0 {
						objects[6].InvFirstItem = objects[c.Input.Chain[0]]
					}
					for at, slot := range c.Input.Chain {
						if at+1 < len(c.Input.Chain) {
							objects[slot].InvNextItem = objects[c.Input.Chain[at+1]]
						}
					}
					if words() != c.Before || c.Before != c.Input.Fields {
						t.Fatal("canonical original fields differ before execution")
					}
					copy(expected, raw)
					for _, write := range c.Writes {
						if write.Slot < 0 || write.Slot >= len(objects) {
							t.Fatal("unowned original flags write")
						}
						off := 0x1000 + write.Slot*0x1000 + placement + int(offsets[2])
						if binary.LittleEndian.Uint32(expected[off:]) != write.Before {
							t.Fatal("original flags write prefix differs")
						}
						binary.LittleEndian.PutUint32(expected[off:], write.After)
					}
					var owner *server.Object
					if c.Input.Owner != -1 {
						if c.Input.Owner != 6 {
							t.Fatal("unexpected canonical owner")
						}
						owner = objects[c.Input.Owner]
					}
					if result := npcEquipmentNativeCall(route, owner); uint32(result) != c.Return || words() != c.After {
						t.Fatal("original return/final fields differ", result, words(), c.Return, c.After)
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("unexpected write in whole64KiB foreign linked owner")
					}
				})
			}
		}
	}
}
