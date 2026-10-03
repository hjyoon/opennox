package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Direct Go body with the production runtime, not an unmodified C dispatcher
// or a naturally advancing stock-map frame. The only wrapped service is real
// BuffOff: after calling it, advance the server's actual Frame DWORD. Native
// protection, sound, update/health records and UnitSetHP remain unchanged.
func TestDefaultDamageLiveFrameNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(9) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, frame := range []uint32{0, 0x80000004, math.MaxUint32} {
		t.Run(fmt.Sprintf("frame-%#x", frame), func(t *testing.T) {
			srv.SetFrame(1400)
			target, source, weapon := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
			update, freeUpdate := alloc.New(server.MonsterUpdateData{})
			playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			health, freeHealth := alloc.New(server.HealthData{})
			initData, freeInit := alloc.New(server.ModifierInitData{})
			for _, free := range []func(){freeUpdate, freePlayerUpdate, freePlayer, freeHealth, freeInit} {
				t.Cleanup(free)
			}
			*update = server.MonsterUpdateData{StatusFlags: 0x40000000, Field547: 99, Field546: 88}
			*player = server.Player{PlayerInd: 7}
			*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			*initData = server.ModifierInitData{}
			target.TypeInd, target.ObjClass, target.ObjFlags, target.ObjSubClass = 71, object.ClassMonster, 0, 0x10
			target.UpdateData, target.HealthData = unsafe.Pointer(update), health
			target.Field131, target.Frame134 = 99, 77
			source.TypeInd, source.ObjClass, source.ObjFlags = 72, object.ClassPlayer, 0
			source.UpdateData, source.PrevPos = unsafe.Pointer(playerUpdate), types.Ptf(-3, 7)
			weapon.TypeInd, weapon.ObjClass, weapon.ObjFlags, weapon.ObjSubClass = 777, object.ClassWand, 0, 0
			weapon.InitData = unsafe.Pointer(initData)
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(update), unsafe.Pointer(playerUpdate), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(initData)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want >4 GiB", ptr)
				}
			}
			r := defaultDamageWorldRuntime4E0B30(srv)
			buffOff, calls := r.BuffOff, 0
			r.BuffOff = func(owner *server.Object, id server.EnchantID) {
				if owner != target || id != server.ENCHANT_INVISIBLE || update.Field547 != 1 || update.Field546 != 777 ||
					target.Pos132 != source.PrevPos || target.Field131 != 99 || target.Frame134 != 77 || srv.Frame() != 1400 {
					t.Fatal("frame advancement ran outside the original hit prefix")
				}
				buffOff(owner, id)
				calls++
				srv.SetFrame(frame)
			}
			if !server.DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r) || calls != 1 || health.Cur != 17 ||
				target.Obj130 != weapon || target.Field131 != uint32(object.DamageBlade) || target.Frame134 != frame || target.Field38 != math.MaxUint32 ||
				update.StatusFlags != 0x40000000|object.MonStatusInjured || update.Field547 != 1 || update.Field546 != 777 {
				t.Fatalf("native live Frame: HP=%d frame=%#x want=%#x calls=%d latch=%d", health.Cur, target.Frame134, frame, calls, update.Field547)
			}
			t.Logf("C-owned >4GiB production-runtime Frame=%#x target=%p weapon=%p HP=20->17", frame, target, weapon)
		})
	}
}
