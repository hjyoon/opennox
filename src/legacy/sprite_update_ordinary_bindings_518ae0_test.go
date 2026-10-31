package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type spriteOrdinaryServer518AE0 struct {
	Server
	srv     *server.Server
	special func(*server.Object, *server.Object) bool
}

func (s *spriteOrdinaryServer518AE0) S() *server.Server { return s.srv }
func (s *spriteOrdinaryServer518AE0) NetUpdateObjectSpecialNative527E50(recipient, unit *server.Object) bool {
	return s.special(recipient, unit)
}

func TestSpriteUpdateOrdinaryBindings518AE0(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	rows, _ := spriteLoad518AE0(t)
	selected := make([]spriteRow518AE0, 0, 10)
	for branch := 0; branch < 8; branch++ {
		found := false
		for _, row := range rows {
			if row.Name == fmt.Sprintf("complete-state-byte-%d", branch) && row.Input.Value == 37 {
				selected = append(selected, row)
				found = true
				break
			}
		}
		if !found {
			t.Fatal("missing original ordinary branch", branch)
		}
	}
	for _, caches := range [][3]uint32{{0, 0, 0}, {117, 117, 117}} {
		found := false
		for _, row := range rows {
			if row.Name == "exact-DWORD-caches-and-precedence" && row.Input.Caches == caches && row.Input.Class == 0 && ((caches[0] == 0 && row.Input.Type == 0) || (caches[0] == 117 && row.Input.Type == 117)) {
				selected = append(selected, row)
				found = true
				break
			}
		}
		if !found {
			t.Fatal("missing original exact-cache branch", caches)
		}
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.NetList = netlist.New()
	srv.NetList.Init()
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if e := srv.Types.ReadObjectType(&things.Thing{Name: "SpriteOrdinaryNative", Class: []things.Class{"IMMOBILE"}, Flags: []things.Flag{"NO_COLLIDE"}}); e != nil {
		t.Fatal(e)
	}
	if !srv.Objs.Init(8) {
		t.Fatal("native object allocator")
	}
	unit := srv.NewObjectByTypeID("SpriteOrdinaryNative")
	linked := srv.NewObjectByTypeID("SpriteOrdinaryNative")
	recipient := srv.NewObjectByTypeID("SpriteOrdinaryNative")
	if unit == nil || linked == nil || recipient == nil {
		t.Fatal("native ordinary objects")
	}
	f := &spriteFixture518AE0{t: t}
	data := spriteAllocate518AE0[[64]byte](t, f)
	other := spriteAllocate518AE0[[64]byte](t, f)
	t.Cleanup(func() { unit.UpdateData = nil; linked.UpdateData = nil; recipient.UpdateData = nil })
	for i := 0; i < 3; i++ {
		ptr := memmap.PtrUint32(0x5d4594, uintptr(2386972+4*i))
		before := *ptr
		t.Cleanup(func() { *ptr = before })
	}
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantListHeadsC(0, 0)
	setImportantAllocClassC(nil)
	setImportantRecipientMaskC(0)
	owner := GetServer
	var fallbackCalls int
	GetServer = func() Server {
		return &spriteOrdinaryServer518AE0{Server: owner(), srv: srv, special: func(a, b *server.Object) bool {
			if a != recipient || b != unit {
				t.Fatal("ordinary fallback lost full native pointer")
			}
			fallbackCalls++
			return true
		}}
	}
	t.Cleanup(func() { GetServer = owner })
	for index, row := range selected {
		t.Run(fmt.Sprintf("%d/%s", index, row.Name), func(t *testing.T) {
			for _, native := range []bool{true, false} {
				t.Run(fmt.Sprintf("C=%t", native), func(t *testing.T) {
					c := row.Input
					unit.TypeInd = c.Type
					unit.ObjClass = object.Class(c.Class)
					unit.ObjSubClass = object.SubClass(c.Subclass)
					unit.ObjFlags = object.Flags(c.Flags)
					unit.NetCode = c.Primary
					unit.Extent = c.Secondary
					data.Value = [64]byte{}
					other.Value = [64]byte{}
					for _, off := range []int{0, 12, 16, 20} {
						binary.LittleEndian.PutUint32(data.Value[off:], c.Value)
					}
					binary.LittleEndian.PutUint32(other.Value[16:], c.Value^0xa5)
					unit.UpdateData = unsafe.Pointer(&data.Value)
					linked.UpdateData = unsafe.Pointer(&other.Value)
					unit.SetElevatorLinkFor(unit.UpdateData, linked)
					if unit.ElevatorLinkFor(unit.UpdateData) != linked {
						t.Fatal("ordinary native sidecar did not retain linked object")
					}
					binary.LittleEndian.PutUint32(data.Value[4:], 0xdeadbeef)
					for i, v := range c.Caches {
						*memmap.PtrUint32(0x5d4594, uintptr(2386972+4*i)) = v
					}
					srv.NetList.ResetAll()
					fallbackCalls = 0
					beforeUnit := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))...)
					beforeLinked := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(linked)), int(unsafe.Sizeof(*linked)))...)
					beforeData, beforeOther := *data, *other
					globals := append([]byte(nil), unsafe.Slice((*byte)(memmap.PtrOff(0x5d4594, 2386972-64)), 140)...)
					var got int32
					if native {
						got = spriteUpdatePublicC518AE0(recipient, int32(c.Recipient), unit)
					} else {
						got = NetSpriteUpdate518AE0(recipient, int32(c.Recipient), unit)
					}
					if uint32(got) != row.Output.Returns[0] {
						t.Fatalf("ordinary return=%#x original=%#x", uint32(got), row.Output.Returns[0])
					}
					publication := row.Output.Events[len(row.Output.Events)-1]
					want, _ := hex.DecodeString(fmt.Sprint(publication["packet"]))
					switch publication["service"] {
					case "0x4e5390":
						player, packet := complexObjectTakeImportantPacketC518960()
						if uint32(player) != c.Recipient || !bytes.Equal(packet, want) {
							t.Fatalf("actual important packet=%x player=%d original=%x", packet, player, want)
						}
					case "0x40ebc0":
						if packet := srv.NetList.CopyPacketsA(ntype.PlayerInd(c.Recipient), netlist.Kind1); !bytes.Equal(packet, want) {
							t.Fatalf("actual ordinary kind1 packet=%x original=%x", packet, want)
						}
					case "0x527e50":
						if fallbackCalls != 1 {
							t.Fatal("ordinary special callback count", fallbackCalls)
						}
					default:
						t.Fatal("unknown original ordinary service", publication)
					}
					first, last := importantNativeListC()
					if first != nil || last != nil {
						t.Fatal("extra ordinary important packet")
					}
					if !bytes.Equal(beforeUnit, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) || !bytes.Equal(beforeLinked, unsafe.Slice((*byte)(unsafe.Pointer(linked)), int(unsafe.Sizeof(*linked)))) || beforeData != *data || beforeOther != *other || !bytes.Equal(globals, unsafe.Slice((*byte)(memmap.PtrOff(0x5d4594, 2386972-64)), 140)) {
						t.Fatal("ordinary full object/update/global-neighbor preservation")
					}
				})
			}
		})
	}
}
