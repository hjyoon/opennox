package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Actual C dispatcher -> registered DefaultDamage -> 004E1320 stock effects
// -> production UnitSetHP. All native records are C-owned, with no effect,
// sound, protection or HP callback substituted by the test. Install only the
// root package's required hurt-state binding; these hits never reach it.
func TestDefaultDamageLateDefendNative4E0B30CDispatcher(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeDefaultLateDefend", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold hit unexpectedly requested a hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	if !srv.Objs.Init(6) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, playerTarget := range []bool{false, true} {
		t.Run(map[bool]string{false: "NPC", true: "player"}[playerTarget], func(t *testing.T) {
			target, source, item := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
			playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
			monsterUpdate, freeMonsterUpdate := alloc.New(server.MonsterUpdateData{})
			sourceUpdate, freeSourceUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			health, freeHealth := alloc.New(server.HealthData{})
			initData, freeInit := alloc.New(server.ModifierInitData{})
			first, freeFirst := alloc.New(server.ModifierEff{})
			second, freeSecond := alloc.New(server.ModifierEff{})
			for _, free := range []func(){freePlayerUpdate, freeMonsterUpdate, freeSourceUpdate, freePlayer, freeHealth, freeInit, freeFirst, freeSecond} {
				t.Cleanup(free)
			}
			*player = server.Player{PlayerInd: 7}
			*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
			*sourceUpdate = *playerUpdate
			*monsterUpdate = server.MonsterUpdateData{}
			*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
			*initData = server.ModifierInitData{}
			target.ObjClass, target.ObjFlags, target.ObjSubClass = object.ClassMonster, 0, 0x10
			target.UpdateData = unsafe.Pointer(monsterUpdate)
			if playerTarget {
				target.ObjClass, target.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUpdate)
			}
			target.HealthData, target.Damage, target.InvFirstItem = health, srv.Types.ByID("NativeDefaultLateDefend").Damage, item
			source.ObjClass, source.ObjFlags, source.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(sourceUpdate)
			// Only flags gate the helper: this item deliberately has no class,
			// health, update data or registered damage callback.
			item.ObjClass, item.ObjFlags, item.InitData = 0, object.FlagEquipped, unsafe.Pointer(initData)
			fns := playerDamageLateDefendFunctionsNative4E1320()
			*first = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.armorMultiplier, Valf: 0.5}}
			*second = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.grip}, DefendCollide88: server.ModifierEffFnc{Val: 0}}
			initData.Modifiers[2], initData.Modifiers[3] = first, second
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(item), unsafe.Pointer(playerUpdate), unsafe.Pointer(monsterUpdate), unsafe.Pointer(sourceUpdate), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(initData), unsafe.Pointer(first), unsafe.Pointer(second)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer=%p, want >4 GiB", ptr)
				}
			}
			if !objectDamageDispatchCallNative(target, source, nil, 7, object.DamageClaw) || health.Cur != 19 ||
				target.Obj130 != source || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageClaw) || target.Field38 != math.MaxUint32 {
				t.Fatalf("C DefaultDamage ordered effects: HP=%d attribution=%p frame=%d", health.Cur, target.Obj130, target.Frame134)
			}
			second.Defend76.Fnc, second.DefendCollide88.Val = fns.inversion, 0
			if !objectDamageDispatchCallNative(target, source, nil, 7, object.DamageClaw) || health.Cur != 19 {
				t.Fatalf("late zero skipped/clamped: HP=%d", health.Cur)
			}
			t.Logf("C dispatcher retained >4GiB DefaultDamage records: player=%t target=%p item=%p first=%p second=%p HP=20->19->19", playerTarget, target, item, first, second)
		})
	}
}
