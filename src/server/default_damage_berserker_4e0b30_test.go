package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// PlayerCollide passes the Warrior as both source and weapon, with CRUSH.
// GAME.EXE 004E0ED0 takes the same-object path, then 004E0FC9 records the
// damage type (not an equipped weapon type) before the normal damage tail.
func TestDefaultDamageWorld4E0B30BerserkerCharge(t *testing.T) {
	for _, subclass := range []uint32{0x202, 0x10} {
		t.Run(object.SubClass(subclass).String(), func(t *testing.T) {
			update := &MonsterUpdateData{Field547: 99, Field546: 999}
			target := &Object{
				ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(subclass),
				UpdateData: unsafe.Pointer(update), HealthData: &HealthData{Cur: 2000, Max: 2000},
				Buffs: 1<<defaultDamageInvisibleEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30 | 1<<defaultDamageShieldEnchant4E0B30,
			}
			source := &Object{TypeInd: 713, ObjClass: object.ClassPlayer, PrevPos: types.Ptf(45, 76)}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source))>>32 == 0 {
				t.Fatal("charge test must exercise a native-width source pointer")
			}
			var events []string
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame: func() uint32 { return 600 }, GameplayFlag1: func() bool { return true },
				IsEnemy: func(_, _ *Object) bool { return true },
				BuffOff: func(got *Object, enchant EnchantID) {
					if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 || update.Field547 != 0 {
						t.Fatalf("BuffOff(%p, %d), hit latch=%d", got, enchant, update.Field547)
					}
					got.Buffs &^= 1 << enchant
					events = append(events, "buff-off")
				},
				DefaultDamageSound: func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != source {
						t.Fatalf("sound target/source=%p/%p", gotTarget, gotSource)
					}
					events = append(events, "sound")
				},
				AdjustFieldGuide: func(gotSource, gotTarget *Object, damage int32) int32 {
					if gotSource != source || gotTarget != target || damage != 150 {
						t.Fatalf("guide source/target/damage=%p/%p/%d", gotSource, gotTarget, damage)
					}
					events = append(events, "field-guide")
					return 180
				},
				ShieldReduce: func(got *Object, damage *int32, typ object.DamageType, attacker *Object) {
					if got != target || *damage != 180 || typ != object.DamageCrush || attacker != source {
						t.Fatalf("shield target/damage/type/source=%p/%d/%d/%p", got, *damage, typ, attacker)
					}
					*damage = 120
					events = append(events, "shield")
				},
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 120 {
						t.Fatalf("damage target/amount=%p/%d", got, damage)
					}
					got.HealthData.Cur -= uint16(damage)
					events = append(events, "damage")
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("Berserker Charge rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, source, 150, object.DamageCrush, runtime) {
				t.Fatal("charge returned false")
			}
			if target.HealthData.Cur != 1880 || target.Pos132 != source.PrevPos || target.Obj130 != source ||
				target.Field131 != uint32(object.DamageCrush) || target.Frame134 != 600 ||
				update.Field547 != 2 || update.Field546 != uint32(object.DamageCrush) || !update.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatalf("charge metadata hp=%d pos=%v source=%p type/frame=%d/%d latch=%d/%d status=%#x",
					target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134,
					update.Field547, update.Field546, update.StatusFlags)
			}
			// sub_4E1400 rejects a PLAYER-class weapon: charge must not trigger
			// the melee Shock retaliation path, even when the target has Shock.
			if !target.HasEnchant(defaultDamageShockEnchant4E0B30) || target.HasEnchant(defaultDamageInvisibleEnchant4E0B30) {
				t.Fatalf("charge buff state=%#x", target.Buffs)
			}
			if want := []string{"buff-off", "sound", "field-guide", "shield", "damage"}; !reflect.DeepEqual(events, want) {
				t.Fatalf("events=%v, want %v", events, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30BerserkerChargeEarlyExits(t *testing.T) {
	for _, tc := range []struct {
		name             string
		gameplay, enemy  bool
		flags            object.Flags
		invulnerable     bool
		shieldAbsorbsAll bool
		wantDamage       bool
		wantReturn       bool
	}{
		{name: "regular enemy", gameplay: true, enemy: true, wantDamage: true, wantReturn: true},
		{name: "regular friendly passes non-melee gate", gameplay: true, wantDamage: true, wantReturn: true},
		{name: "campaign enemy", enemy: true, wantDamage: true, wantReturn: true},
		{name: "campaign friendly owner gate", wantReturn: true},
		{name: "NoUpdate remains immune", gameplay: true, enemy: true, flags: object.FlagNoUpdate, wantReturn: true},
		{name: "dead remains immune", gameplay: true, enemy: true, flags: object.FlagDead, wantReturn: true},
		{name: "invulnerable remains immune", gameplay: true, enemy: true, invulnerable: true, wantReturn: true},
		{name: "shield fully absorbs charge", gameplay: true, enemy: true, shieldAbsorbsAll: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			update := &MonsterUpdateData{}
			target := &Object{ObjClass: object.ClassMonster, ObjFlags: tc.flags,
				UpdateData: unsafe.Pointer(update), HealthData: &HealthData{Cur: 2000, Max: 2000}}
			source := &Object{ObjClass: object.ClassPlayer}
			if tc.invulnerable {
				target.Buffs |= 1 << defaultDamageInvulnerableEnchant4E0B30
			}
			if tc.shieldAbsorbsAll {
				target.Buffs |= 1 << defaultDamageShieldEnchant4E0B30
			}
			called := false
			runtime := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return tc.gameplay },
				IsEnemy:       func(_, _ *Object) bool { return tc.enemy },
				ShieldReduce:  func(_ *Object, damage *int32, _ object.DamageType, _ *Object) { *damage = 0 },
				DamageClear: func(got *Object, damage int32) {
					called = true
					if got != target || damage != 150 {
						t.Fatalf("DamageClear(%p,%d), want (%p,150)", got, damage, target)
					}
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("charge rejected: %s", reason)
				},
			}
			if got := DefaultDamageWorld4E0B30(target, source, source, 150, object.DamageCrush, runtime); got != tc.wantReturn || called != tc.wantDamage {
				t.Fatalf("return/damage=%t/%t, want %t/%t", got, called, tc.wantReturn, tc.wantDamage)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30BerserkerChargeNPCDefense(t *testing.T) {
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	init := &ModifierInitData{}
	init.Modifiers[2] = modifier
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
	target := &Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10,
		UpdateData: unsafe.Pointer(&MonsterUpdateData{}), HealthData: &HealthData{Cur: 2000, Max: 2000}, InvFirstItem: armor}
	source := &Object{ObjClass: object.ClassPlayer}
	var events []string
	if !DefaultDamageWorld4E0B30(target, source, source, 150, object.DamageCrush, DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		CanApplyLateDefend: func(got *ModifierEff) bool {
			return got == modifier
		},
		BuffOff: func(_ *Object, _ EnchantID) { events = append(events, "buff-off") },
		ApplyLateDefend: func(got *ModifierEff, item, victim, weapon, attacker *Object, damage int32, typ object.DamageType) int32 {
			if got != modifier || item != armor || victim != target || weapon != source || attacker != source || damage != 150 || typ != object.DamageCrush {
				t.Fatalf("wrong charge defense arguments: %p/%p/%p/%p/%p/%d/%d", got, item, victim, weapon, attacker, damage, typ)
			}
			events = append(events, "defend")
			return 100
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 100 {
				t.Fatalf("DamageClear(%p,%d), want defended charge 100", got, damage)
			}
			got.HealthData.Cur -= uint16(damage)
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("NPC charge rejected: %s", reason)
		},
	}) {
		t.Fatal("defended charge returned false")
	}
	if target.HealthData.Cur != 1900 || !reflect.DeepEqual(events, []string{"buff-off", "defend", "damage"}) {
		t.Fatalf("defended charge hp=%d events=%v", target.HealthData.Cur, events)
	}
}

func TestDefaultDamageWorld4E0B30BerserkerChargeAdmissionBoundary(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	otherPlayer := &Object{ObjClass: object.ClassPlayer}
	monster := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{})}
	weapon := &Object{ObjClass: object.ClassWeapon}
	for _, tc := range []struct {
		name           string
		source, weapon *Object
		typ            object.DamageType
	}{
		{"source-less crush", nil, nil, object.DamageCrush},
		{"player without weapon", player, nil, object.DamageCrush},
		{"different player weapon", player, otherPlayer, object.DamageCrush},
		{"equipped weapon crush", player, weapon, object.DamageCrush},
		{"monster self crush", monster, monster, object.DamageCrush},
		{"player self blade", player, player, object.DamageBlade},
		{"player self bite", player, player, object.DamageBite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{ObjClass: object.ClassMonster, HealthData: &HealthData{Cur: 2000, Max: 2000},
				UpdateData: unsafe.Pointer(&MonsterUpdateData{})}
			rejected := false
			DefaultDamageWorld4E0B30(target, tc.source, tc.weapon, 150, tc.typ, DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true },
				IsEnemy:       func(_, _ *Object) bool { return true },
				DamageClear:   func(_ *Object, _ int32) { t.Fatal("unported damage shape was admitted") },
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					rejected = reason == "unsupported monster damage shape"
				},
			})
			if !rejected {
				t.Fatal("unported damage shape was not reported")
			}
		})
	}
}
