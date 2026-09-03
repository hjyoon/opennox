package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func durabilityItem4E1560(class object.Class, subclass uint32, health uint16) (
	*Object, *HealthData, *WeaponArmorUpdateData, *ModifierInitData,
) {
	hp := &HealthData{Cur: health, Max: health}
	update := &WeaponArmorUpdateData{}
	init := &ModifierInitData{}
	item := &Object{
		ObjClass:    class,
		ObjSubClass: object.SubClass(subclass),
		HealthData:  hp,
		UpdateData:  unsafe.Pointer(update),
		InitData:    unsafe.Pointer(init),
		Damage:      unsafe.Pointer(new(byte)),
	}
	return item, hp, update, init
}

func TestPlayerDamageWeaponNative4E1560NoOpGates(t *testing.T) {
	called := false
	runtime := ItemDurabilityDamageRuntime4E1560{
		Damage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
			called = true
			return true
		},
	}
	if !PlayerDamageWeaponNative4E1560(nil, nil, nil, nil, 7, object.DamageBlade, runtime) {
		t.Fatal("nil item was not handled")
	}
	item := &Object{ObjClass: object.ClassWeapon}
	if !PlayerDamageWeaponNative4E1560(item, nil, nil, nil, 7, object.DamageBlade, runtime) {
		t.Fatal("nil health was not handled")
	}

	item, health, update, _ := durabilityItem4E1560(object.ClassWeapon, 0x00800000, 9)
	if !PlayerDamageWeaponNative4E1560(item, nil, nil, nil, 7, object.DamageBlade, runtime) {
		t.Fatal("excluded weapon was not handled")
	}
	if called || health.Cur != 9 || update.Field0 != 0 {
		t.Fatalf("excluded weapon mutated state: called=%t health=%d carry=%#x", called, health.Cur, update.Field0)
	}
}

func TestPlayerDamageWeaponNative4E1560CarriesAndRoundsToEven(t *testing.T) {
	item, health, update, _ := durabilityItem4E1560(object.ClassWeapon, 0, 5)
	owner := &Object{ObjClass: object.ClassPlayer}
	source := &Object{}
	effective := &Object{}
	var damages []int32
	var reports [][2]uint16
	runtime := ItemDurabilityDamageRuntime4E1560{
		Damage: func(gotItem, gotSource, gotEffective *Object, damage int32, typ object.DamageType) bool {
			if gotItem != item || gotSource != source || gotEffective != effective || typ != object.DamageBlade {
				t.Fatalf("damage args = %p/%p/%p/%d", gotItem, gotSource, gotEffective, typ)
			}
			damages = append(damages, damage)
			health.Cur -= uint16(damage)
			return true
		},
		ReportHealth: func(gotOwner, gotItem *Object, before, after uint16) {
			if gotOwner != owner || gotItem != item {
				t.Fatalf("report args = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
			}
			reports = append(reports, [2]uint16{before, after})
		},
	}

	if !PlayerDamageWeaponNative4E1560(item, owner, source, effective, 0.5, object.DamageBlade, runtime) {
		t.Fatal("first half point was not handled")
	}
	if len(damages) != 0 || health.Cur != 5 || math.Float32frombits(update.Field0) != 0.5 {
		t.Fatalf("first half point = damages:%v health:%d carry:%g", damages, health.Cur, math.Float32frombits(update.Field0))
	}
	if !PlayerDamageWeaponNative4E1560(item, owner, source, effective, 0.5, object.DamageBlade, runtime) {
		t.Fatal("second half point was not handled")
	}
	if len(damages) != 1 || damages[0] != 1 || health.Cur != 4 || math.Float32frombits(update.Field0) != 0 {
		t.Fatalf("second half point = damages:%v health:%d carry:%g", damages, health.Cur, math.Float32frombits(update.Field0))
	}
	if len(reports) != 1 || reports[0] != [2]uint16{5, 4} {
		t.Fatalf("health reports = %v, want [[5 4]]", reports)
	}
}

func TestItemDurabilityRound4E1560MatchesX87Conversion(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{value: -1.5, want: -2},
		{value: -0.5, want: 0},
		{value: 0.5, want: 0},
		{value: 1.5, want: 2},
		{value: 2.5, want: 2},
		{value: float32(math.Inf(1)), want: math.MinInt32},
		{value: float32(math.Inf(-1)), want: math.MinInt32},
		{value: float32(math.NaN()), want: math.MinInt32},
		{value: 2147483648, want: math.MinInt32},
	}
	for _, tc := range tests {
		if got := itemDurabilityRound4E1560(tc.value); got != tc.want {
			t.Fatalf("itemDurabilityRound4E1560(%g) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestItemDurabilityDamageNative4E1560ModifierOrderAndEquipRules(t *testing.T) {
	item, health, update, init := durabilityItem4E1560(object.ClassWeapon, 0x00800000, 8)
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte)), Valf: 1.25}}
	init.Modifiers[1] = modifier
	owner := &Object{}
	source := &Object{}
	effective := &Object{}
	var damage int32
	runtime := ItemDurabilityDamageRuntime4E1560{
		ApplyDefend: func(got *ModifierEff, gotItem, gotOwner, gotEffective, gotSource *Object, value *float32) bool {
			if got != modifier || gotItem != item || gotOwner != owner || gotEffective != effective || gotSource != source {
				t.Fatalf("modifier args lost original order")
			}
			*value *= 2
			return true
		},
		Damage: func(gotItem, gotSource, gotEffective *Object, gotDamage int32, typ object.DamageType) bool {
			if gotItem != item || gotSource != source || gotEffective != effective || typ != object.DamageCrush {
				t.Fatalf("damage args lost original order")
			}
			damage = gotDamage
			health.Cur -= uint16(gotDamage)
			return true
		},
	}
	update.Field0 = math.Float32bits(0.25)
	if !EquipDamageNative4E16D0(item, owner, source, effective, 0.25, object.DamageCrush, runtime) {
		t.Fatal("equipped-item damage was not handled")
	}
	if damage != 1 || health.Cur != 7 || math.Float32frombits(update.Field0) != -0.25 {
		t.Fatalf("equipped-item damage = damage:%d health:%d carry:%g", damage, health.Cur, math.Float32frombits(update.Field0))
	}
}

func TestItemDurabilityDamageNative4E1560FailsClosedBeforeCarryMutation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Object, *ModifierInitData)
	}{
		{name: "nil update", mutate: func(item *Object, _ *ModifierInitData) { item.UpdateData = nil }},
		{name: "nil init", mutate: func(item *Object, _ *ModifierInitData) { item.InitData = nil }},
		{name: "missing damage callback", mutate: func(item *Object, _ *ModifierInitData) { item.Damage = nil }},
		{name: "unsupported modifier", mutate: func(_ *Object, init *ModifierInitData) {
			init.Modifiers[1] = &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, health, update, init := durabilityItem4E1560(object.ClassArmor, 0, 6)
			update.Field0 = math.Float32bits(0.25)
			tc.mutate(item, init)
			unsupported := 0
			runtime := ItemDurabilityDamageRuntime4E1560{
				ApplyDefend: func(*ModifierEff, *Object, *Object, *Object, *Object, *float32) bool { return false },
				Damage:      func(*Object, *Object, *Object, int32, object.DamageType) bool { return true },
				Unsupported: func(string, *Object, *Object, *Object, *Object, float32, object.DamageType) {
					unsupported++
				},
			}
			if EquipDamageNative4E16D0(item, nil, nil, nil, 1, object.DamageBlade, runtime) {
				t.Fatal("unsupported branch reported handled")
			}
			if unsupported != 1 || health.Cur != 6 || math.Float32frombits(update.Field0) != 0.25 {
				t.Fatalf("failed-closed state = unsupported:%d health:%d carry:%g", unsupported, health.Cur, math.Float32frombits(update.Field0))
			}
		})
	}
}
