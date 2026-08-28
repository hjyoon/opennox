package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerDamageFixture4E17B0(t *testing.T) (*Object, *Object, unsafe.Pointer) {
	t.Helper()
	sound := unsafe.Pointer(new(byte))
	player := &Player{ArmorEquip: 0x405, WeaponEquip: 0x100}
	update := &PlayerUpdateData{Player: player, Field57: math.Float32bits(0.03), State: PlayerState13}
	target := &Object{
		ObjClass:    object.ClassPlayer,
		ObjFlags:    object.FlagActive | object.FlagEnabled,
		UpdateData:  unsafe.Pointer(update),
		HealthData:  &HealthData{Cur: 20, Field2: 20, Max: 20},
		DamageSound: sound,
	}
	source := monsterActionTestObject50A910(t)
	source.ObjClass = object.ClassMonster
	source.PrevPos.X = 44
	return target, source, sound
}

func playerDamageRuntime4E17B0(t *testing.T, sound unsafe.Pointer, damages *[]int32) PlayerDamageRuntime4E17B0 {
	t.Helper()
	return PlayerDamageRuntime4E17B0{
		Frame:              func() uint32 { return 700 },
		QuestMode:          func() bool { return false },
		QuestDamageScale:   func() float32 { return 1 },
		GodMode:            func() bool { return false },
		IsEnemy:            func(*Object, *Object) bool { return true },
		BuffOff:            func(*Object, EnchantID) {},
		ItemArmorValue:     func(*Object) float32 { return 0.01 },
		FireProtection:     func(*Object) float32 { return 0 },
		PlayerDamageSoundC: sound,
		DamageClear: func(target *Object, damage int32) {
			*damages = append(*damages, damage)
			if int32(target.HealthData.Cur) <= damage {
				target.HealthData.Cur = 0
			} else {
				target.HealthData.Cur -= uint16(damage)
			}
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("unexpected unsupported branch: %s", reason)
		},
	}
}

func TestPlayerDamageNative4E17B0SourceLessLava(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	target.Pos132.X = 99
	var damages []int32
	var events []string
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.QuestMode = func() bool { return true }
	runtime.QuestDamageScale = func() float32 {
		events = append(events, "quest-scale")
		return 0.5
	}
	runtime.FireProtection = func(got *Object) float32 {
		if got != target {
			t.Fatalf("FireProtection(%p), want %p", got, target)
		}
		events = append(events, "fire-protection")
		return 0.25
	}
	runtime.Audio = func(id int, got *Object) {
		if id != 104 || got != target {
			t.Fatalf("Audio(%d,%p), want (104,%p)", id, got, target)
		}
		events = append(events, "fire-sound")
	}
	runtime.BuffOff = func(got *Object, enchant EnchantID) {
		if got != target || enchant != playerDamageInvisibleEnchant4E17B0 {
			t.Fatalf("BuffOff(%p,%d)", got, enchant)
		}
		events = append(events, "buff-off")
	}
	runtime.PlayerDamageSound = func(gotTarget, gotSource *Object) {
		if gotTarget != target || gotSource != nil {
			t.Fatalf("PlayerDamageSound(%p,%p), want (%p,nil)", gotTarget, gotSource, target)
		}
		events = append(events, "damage-sound")
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 5, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("source-less LAVA = handled:%t result:%t", handled, result)
	}
	if len(damages) != 1 || damages[0] != 2 {
		t.Fatalf("LAVA damages = %v, want [2]", damages)
	}
	wantEvents := []string{"quest-scale", "fire-protection", "fire-sound", "buff-off", "damage-sound"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	update := target.UpdateDataPlayer()
	if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamageLava)) ||
		target.Pos132 != (types.Pointf{}) || target.Obj130 != nil || target.Field131 != uint32(object.DamageLava) || target.Frame134 != 700 {
		t.Fatalf("LAVA metadata = marker:%#x/%#x pos:%v source:%p type:%d frame:%d",
			update.Field75, update.Field76, target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
}

func TestPlayerDamageNative4E17B0LavaDamagesEquippedArmor(t *testing.T) {
	target, _, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = math.Float32bits(0.5)
	carry := float32(0.4)
	armor := &Object{
		ObjClass:   object.ClassArmor,
		ObjFlags:   object.FlagEquipped,
		HealthData: &HealthData{Cur: 10, Max: 10},
		UpdateData: unsafe.Pointer(&carry),
		InitData:   unsafe.Pointer(&ModifierInitData{}),
	}
	target.InvFirstItem = armor
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	runtime.ItemArmorValue = func(got *Object) float32 {
		if got != armor {
			t.Fatalf("ItemArmorValue(%p), want %p", got, armor)
		}
		return 0.5
	}
	runtime.CanDamageArmor = func(got *Object) bool { return got == armor }
	var reported bool
	runtime.DamageArmor = func(got, source, weapon *Object, damage int32, typ object.DamageType) bool {
		if got != armor || source != nil || weapon != nil || damage != 2 || typ != object.DamageLava {
			t.Fatalf("DamageArmor(%p,%p,%p,%d,%d)", got, source, weapon, damage, typ)
		}
		got.HealthData.Cur -= uint16(damage)
		return true
	}
	runtime.ReportArmorHealth = func(owner, got *Object, before, after uint16) {
		if owner != target || got != armor || before != 10 || after != 8 {
			t.Fatalf("ReportArmorHealth(%p,%p,%d,%d)", owner, got, before, after)
		}
		reported = true
	}
	if handled, result := PlayerDamageNative4E17B0(target, nil, nil, 2, object.DamageLava, runtime); !handled || !result {
		t.Fatalf("armored LAVA = handled:%t result:%t", handled, result)
	}
	wantCarry := float32(float64(float32(0.5)) / float64(float32(0.5)) * float64(int32(2)))
	wantCarry += float32(0.4)
	wantCarry -= 2
	if armor.HealthData.Cur != 8 || math.Float32bits(carry) != math.Float32bits(wantCarry) || !reported {
		t.Fatalf("armor state = hp:%d carry:%v reported:%t", armor.HealthData.Cur, carry, reported)
	}
	if !reflect.DeepEqual(damages, []int32{2}) {
		t.Fatalf("player damages = %v, want [2]", damages)
	}
}

func TestPlayerDamageNative4E17B0SpiderBiteSequence(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	for i := 0; i < 3; i++ {
		carry := new(float32)
		item := &Object{
			ObjClass:    object.ClassArmor,
			ObjFlags:    object.FlagEquipped,
			HealthData:  &HealthData{Cur: 10, Max: 10},
			UpdateData:  unsafe.Pointer(carry),
			InitData:    unsafe.Pointer(&ModifierInitData{}),
			InvNextItem: target.InvFirstItem,
		}
		target.InvFirstItem = item
	}
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	for i := 0; i < 7; i++ {
		handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime)
		if !handled || !result {
			t.Fatalf("hit %d = handled:%t result:%t", i, handled, result)
		}
	}
	want := []int32{3, 3, 3, 3, 3, 2, 3}
	if len(damages) != len(want) {
		t.Fatalf("damages = %v, want %v", damages, want)
	}
	for i := range want {
		if damages[i] != want[i] {
			t.Fatalf("damages = %v, want %v", damages, want)
		}
	}
	if target.HealthData.Cur != 0 || target.Obj130 != source || target.Field131 != uint32(object.DamageBite) ||
		target.Frame134 != 700 || target.Pos132.X != 44 {
		t.Fatalf("final target state = hp:%d source:%p type:%d frame:%d pos:%v",
			target.HealthData.Cur, target.Obj130, target.Field131, target.Frame134, target.Pos132)
	}
	update := target.UpdateDataPlayer()
	if update.Field76 != 2 || update.Field75 != math.Float32bits(float32(object.DamageBite)) {
		t.Fatalf("damage marker = %#x/%#x", update.Field75, update.Field76)
	}
}

func TestPlayerDamageNative4E17B0RejectsLateDefendBeforeMutation(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	marker := unsafe.Pointer(new(byte))
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: marker}}
	target.InvFirstItem = &Object{
		ObjClass: object.ClassWeapon,
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}}),
	}
	beforeTarget := *target
	beforeUpdate := *target.UpdateDataPlayer()
	var reason string
	runtime := playerDamageRuntime4E17B0(t, sound, new([]int32))
	runtime.Unsupported = func(got string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = got }
	handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime)
	if handled || result || reason != "late equipped-item defend effect" {
		t.Fatalf("result = %t/%t reason %q", handled, result, reason)
	}
	if *target != beforeTarget || *target.UpdateDataPlayer() != beforeUpdate {
		t.Fatal("unsupported branch mutated player state")
	}
}

func TestPlayerDamageNative4E17B0EntryGates(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	var damages []int32
	runtime := playerDamageRuntime4E17B0(t, sound, &damages)
	target.ObjFlags |= object.FlagDead
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || result {
		t.Fatalf("dead gate = %t/%t", handled, result)
	}
	target.ObjFlags &^= object.FlagDead
	target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
	if handled, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, runtime); !handled || !result {
		t.Fatalf("invulnerability gate = %t/%t", handled, result)
	}
	if len(damages) != 0 {
		t.Fatalf("entry gate applied damage: %v", damages)
	}
}
