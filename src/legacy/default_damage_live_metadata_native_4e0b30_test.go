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

// This is a direct Go body/production-runtime test, not an unmodified C
// dispatcher or natural stock-map mutation. Only BuffOff is wrapped: call
// the real service, then replace the C-owned update before the injured store.
// All other services, including the final native UnitSetHP, are unchanged.
func TestDefaultDamageLiveMetadataNative4E0B30CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(9) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, latch := range []uint32{0, 7, math.MaxUint32} {
		t.Run(fmt.Sprintf("latch-%#x", latch), func(t *testing.T) {
			target, source, weapon := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
			old, freeOld := alloc.New(server.MonsterUpdateData{})
			live, freeLive := alloc.New(server.MonsterUpdateData{})
			playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			health, freeHealth := alloc.New(server.HealthData{})
			initData, freeInit := alloc.New(server.ModifierInitData{})
			for _, free := range []func(){freeOld, freeLive, freePlayerUpdate, freePlayer, freeHealth, freeInit} {
				t.Cleanup(free)
			}
			*old = server.MonsterUpdateData{StatusFlags: 0x40000000, Field547: 99, Field546: 88}
			*live = server.MonsterUpdateData{StatusFlags: 0x20000000, Field547: latch, Field546: 0x11223344}
			*player = server.Player{PlayerInd: 7}
			*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			*initData = server.ModifierInitData{}
			target.TypeInd, target.ObjClass, target.ObjFlags, target.ObjSubClass = 71, object.ClassMonster, 0, 0x10
			target.UpdateData, target.HealthData = unsafe.Pointer(old), health
			target.Field131, target.Frame134 = 99, 77
			source.TypeInd, source.ObjClass, source.ObjFlags = 72, object.ClassPlayer, 0
			source.UpdateData, source.PrevPos = unsafe.Pointer(playerUpdate), types.Ptf(-3, 7)
			weapon.TypeInd, weapon.ObjClass, weapon.ObjFlags, weapon.ObjSubClass = 777, object.ClassWand, 0, 0
			weapon.InitData = unsafe.Pointer(initData)
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(old), unsafe.Pointer(live), unsafe.Pointer(playerUpdate), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(initData)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want >4 GiB", ptr)
				}
			}
			r := defaultDamageWorldRuntime4E0B30(srv)
			buffOff, calls := r.BuffOff, 0
			r.BuffOff = func(owner *server.Object, id server.EnchantID) {
				if owner != target || id != server.ENCHANT_INVISIBLE || old.Field547 != 1 || old.Field546 != 777 ||
					target.Pos132 != source.PrevPos || target.Field131 != 99 || target.Frame134 != 77 {
					t.Fatal("live update replacement ran outside the original metadata prefix")
				}
				buffOff(owner, id)
				calls++
				target.UpdateData = unsafe.Pointer(live)
			}
			wantLatch, wantKind := latch, uint32(0x11223344)
			if latch == 0 {
				wantLatch, wantKind = 2, uint32(object.DamageBlade)
			}
			if !server.DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r) || calls != 1 || health.Cur != 17 ||
				target.UpdateData != unsafe.Pointer(live) || target.Obj130 != weapon || target.Field131 != uint32(object.DamageBlade) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 ||
				old.StatusFlags != 0x40000000 || old.Field547 != 1 || old.Field546 != 777 ||
				live.StatusFlags != 0x20000000|object.MonStatusInjured || live.Field547 != wantLatch || live.Field546 != wantKind {
				t.Fatalf("live native metadata: HP=%d old flags=%#x live flags=%#x latch=%#x kind=%#x calls=%d", health.Cur, uint32(old.StatusFlags), uint32(live.StatusFlags), live.Field547, live.Field546, calls)
			}
			t.Logf("C-owned >4GiB body/runtime records: target=%p old=%p live=%p weapon=%p HP=20->17 latch=%#x", target, old, live, weapon, live.Field547)
		})
	}
}
