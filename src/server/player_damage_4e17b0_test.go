package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
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
		GodMode:            func() bool { return false },
		IsEnemy:            func(*Object, *Object) bool { return true },
		BuffOff:            func(*Object, EnchantID) {},
		ItemArmorValue:     func(*Object) float32 { return 0.01 },
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
