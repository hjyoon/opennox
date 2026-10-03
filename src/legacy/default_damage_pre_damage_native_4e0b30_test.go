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

// Actual C dispatcher -> registered DefaultDamage -> native 004E13B0 ->
// stock Vampirism/Sympathy -> production UnitSetHP. No effect, protection,
// sound or HP service is replaced. The test supplies only the root package's
// required hurt-state binding, which must not run for this eight-damage hit.
func TestDefaultDamagePreDamageNative4E0B30CDispatcher(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeDefaultPreDamage", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold hit requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	if !srv.Objs.Init(3) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	target, source, weapon := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
	playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
	monsterUpdate, freeMonsterUpdate := alloc.New(server.MonsterUpdateData{})
	player, freePlayer := alloc.New(server.Player{})
	targetHP, freeTargetHP := alloc.New(server.HealthData{})
	sourceHP, freeSourceHP := alloc.New(server.HealthData{})
	initData, freeInit := alloc.New(server.ModifierInitData{})
	first, freeFirst := alloc.New(server.ModifierEff{})
	second, freeSecond := alloc.New(server.ModifierEff{})
	for _, free := range []func(){freePlayerUpdate, freeMonsterUpdate, freePlayer, freeTargetHP, freeSourceHP, freeInit, freeFirst, freeSecond} {
		t.Cleanup(free)
	}
	*player = server.Player{PlayerInd: 7}
	*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
	*monsterUpdate = server.MonsterUpdateData{}
	*targetHP = server.HealthData{Cur: 20, Max: 20, Field2: 20}
	*sourceHP = server.HealthData{Cur: 19, Max: 20, Field2: 20}
	*initData = server.ModifierInitData{}
	target.TypeInd, target.ObjClass, target.ObjFlags = 71, object.ClassPlayer, 0
	target.UpdateData, target.HealthData = unsafe.Pointer(playerUpdate), targetHP
	target.Damage = srv.Types.ByID("NativeDefaultPreDamage").Damage
	source.TypeInd, source.ObjClass, source.ObjFlags, source.ObjSubClass = 72, object.ClassMonster, 0, 0x11012
	source.UpdateData, source.HealthData = unsafe.Pointer(monsterUpdate), sourceHP
	weapon.TypeInd, weapon.ObjClass, weapon.ObjFlags, weapon.ObjSubClass = 444, object.ClassWand, 0, 0
	weapon.InitData = unsafe.Pointer(initData)
	fns := itemPreDamageFunctionsNative4E13B0()
	*first = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.vampirism, Valf: 0.5}}
	*second = server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: fns.sympathy, Valf: 0.25}}
	initData.Modifiers[0], initData.Modifiers[3] = first, second
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), unsafe.Pointer(playerUpdate), unsafe.Pointer(monsterUpdate), unsafe.Pointer(player), unsafe.Pointer(targetHP), unsafe.Pointer(sourceHP), unsafe.Pointer(initData), unsafe.Pointer(first), unsafe.Pointer(second)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want >4 GiB", ptr)
		}
	}
	if !srv.IsEnemyTo(target, source) {
		t.Fatal("native fixture must be hostile")
	}
	if !objectDamageDispatchCallNative(target, source, weapon, 8, object.DamageBlade) || targetHP.Cur != 12 || sourceHP.Cur != 18 ||
		target.Obj130 != weapon || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageBlade) || target.Field38 != math.MaxUint32 ||
		monsterUpdate.Field130 != 1400 || playerUpdate.State != server.PlayerState13 {
		t.Fatalf("actual C pre-Damage: target HP=%d source HP=%d attribution=%p frame=%d source latch=%d", targetHP.Cur, sourceHP.Cur, target.Obj130, target.Frame134, monsterUpdate.Field130)
	}
	t.Logf("C dispatcher retained >4GiB pre-Damage: target=%p source=%p weapon=%p base=%p first=%p second=%p; target HP=20->12, source HP=19->20->18", target, source, weapon, initData, first, second)
}
