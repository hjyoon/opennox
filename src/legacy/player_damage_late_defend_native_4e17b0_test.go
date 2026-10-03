package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Actual C damage dispatcher -> registered PlayerDamage -> ordered 004E1320
// helper -> production stock effects -> UnitSetHP. No effect/HP substitution;
// every object, item, init, modifier, update, player and HP record is C-owned.
// Capture only the sound boundary normally installed by the root package.
func TestPlayerDamageLateDefendNative4E17B0CDispatcher(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(4) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	target := srv.Objs.NewObject(&server.ObjectType{})
	item := srv.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	player, freePlayer := alloc.New(server.Player{})
	health, freeHealth := alloc.New(server.HealthData{})
	initData, freeInit := alloc.New(server.ModifierInitData{})
	first, freeFirst := alloc.New(server.ModifierEff{})
	second, freeSecond := alloc.New(server.ModifierEff{})
	for _, free := range []func(){freeUpdate, freePlayer, freeHealth, freeInit, freeFirst, freeSecond} {
		t.Cleanup(free)
	}
	*player = server.Player{PlayerInd: 7}
	*update = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
	*health = server.HealthData{Cur: 20, Max: 20, Field2: 20}
	*initData = server.ModifierInitData{}
	target.ObjClass, target.ObjFlags = object.ClassPlayer, 0
	target.UpdateData, target.HealthData = unsafe.Pointer(update), health
	target.Damage, target.InvFirstItem = playerDamageMeleeCallbackNative4E17B0(), item
	// Class zero, no item HP/update/damage: only equipped flags gate 004E1320.
	item.ObjClass, item.ObjFlags, item.InitData = 0, object.FlagEquipped, unsafe.Pointer(initData)
	fns := playerDamageLateDefendFunctionsNative4E1320()
	*first = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.armorMultiplier, Valf: 0.5}}
	*second = server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fns.grip}, DefendCollide88: server.ModifierEffFnc{Val: 0}}
	initData.Modifiers[2], initData.Modifiers[3] = first, second
	oldSound := Nox_xxx_soundPlayerDamageSound_5328B0
	t.Cleanup(func() { Nox_xxx_soundPlayerDamageSound_5328B0 = oldSound })
	soundCalls := 0
	Nox_xxx_soundPlayerDamageSound_5328B0 = func(got, source *server.Object) int {
		if got != target || source != nil {
			t.Fatal("player damage sound boundary arguments")
		}
		soundCalls++
		return 0
	}
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(item), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(initData), unsafe.Pointer(first), unsafe.Pointer(second)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want >4 GiB", ptr)
		}
	}
	if !objectDamageDispatchCallNative(target, nil, nil, 7, object.DamagePoison) {
		t.Fatal("registered PlayerDamage rejected poison")
	}
	if health.Cur != 19 || target.Frame134 != 1400 || target.Field131 != uint32(object.DamagePoison) ||
		update.Field76 != 2 || update.Field75 != uint32(object.DamagePoison) || target.Field38 != math.MaxUint32 {
		t.Fatalf("ordered C dispatch: HP=%d marker=%d/%d frame=%d", health.Cur, update.Field76, update.Field75, target.Frame134)
	}
	// A zero overwritten by Inversion must reach the HP tail as zero, not
	// be raised to the original pre-Defend minimum of one.
	second.Defend76.Fnc, second.DefendCollide88.Val = fns.inversion, 0
	if !objectDamageDispatchCallNative(target, nil, nil, 7, object.DamagePoison) || health.Cur != 19 || soundCalls != 2 {
		t.Fatalf("late zero was skipped/clamped: HP=%d", health.Cur)
	}
	t.Logf("C dispatcher retained >4GiB PlayerDamage/item/modifier records: target=%p item=%p first=%p second=%p HP=20->19->19", target, item, first, second)
}
