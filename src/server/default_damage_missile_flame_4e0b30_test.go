package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func damageFlameFixture4E17B0(t *testing.T, fromPlayer, toPlayer, splash bool) (target, source, weapon, missile *Object) {
	t.Helper()
	target, source = damageMeleeUnitFixture4E17B0(t, toPlayer), damageMeleeUnitFixture4E17B0(t, fromPlayer)
	missile = &Object{TypeInd: 695, ObjClass: object.ClassMissile, ObjOwner: source, PrevPos: source.PrevPos.Mul(2)}
	weapon = missile
	if splash {
		source, weapon = missile, nil
	}
	return
}

func TestDefaultDamageWorld4E0B30MissileFlameMatrix(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for _, splash := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/splash-%t", fromPlayer, toPlayer, splash), func(t *testing.T) {
					target, source, weapon, missile := damageFlameFixture4E17B0(t, fromPlayer, toPlayer, splash)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.FireProtection = func(*Object) float64 { return 0.5 }
					if !DefaultDamageWorld4E0B30(target, source, weapon, 9, object.DamageFlame, r) ||
						target.HealthData.Cur != 196 || target.Obj130 != missile || target.Pos132 != missile.PrevPos ||
						target.Field131 != 1 || target.Frame134 != 1400 {
						t.Fatalf("FLAME HP=%d attribution=%p position=%v type/frame=%d/%d", target.HealthData.Cur,
							target.Obj130, target.Pos132, target.Field131, target.Frame134)
					}
					if !toPlayer {
						ud := target.UpdateDataMonster()
						wantMarker, wantType := uint32(1), uint32(missile.TypeInd)
						if splash {
							wantMarker, wantType = 2, 1
						}
						if ud.Field547 != wantMarker || ud.Field546 != wantType ||
							!ud.StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) {
							t.Fatalf("FLAME NPC marker=%d/%d flags=%x", ud.Field547, ud.Field546, ud.StatusFlags)
						}
					}
					if !fromPlayer && source.FindOwnerChainPlayer().UpdateDataMonster().Field130 != 1400 {
						t.Fatal("FLAME NPC owner successful-hit timestamp missing")
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30MissileFlameTailOrder(t *testing.T) {
	target, source, weapon, _ := damageFlameFixture4E17B0(t, false, true, false)
	late := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, late}})}
	target.InvFirstItem, target.Buffs = armor, 1<<defaultDamageShieldEnchant4E0B30
	target.DamageSound = unsafe.Pointer(new(byte))
	frame := uint32(1400)
	var events []string
	r := damageMeleeWorldRuntime4E0B30(t)
	r.Frame = func() uint32 { return frame }
	r.FireProtection = func(*Object) float64 { events = append(events, "resistance"); frame++; return 0.25 }
	r.BuffOff = func(_ *Object, id EnchantID) {
		if id != defaultDamageInvisibleEnchant4E0B30 {
			t.Fatal("wrong buff removal")
		}
		events = append(events, "invisible-off")
		frame++
	}
	r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == late }
	r.ApplyLateDefend = func(_ *ModifierEff, _, _, gotWeapon, gotSource *Object, damage int32, _ object.DamageType) int32 {
		if gotWeapon != weapon || gotSource != source || damage != 6 {
			t.Fatal("late defense did not follow resistance")
		}
		events = append(events, "late-defense")
		frame++
		return damage + 4
	}
	r.PlayerDamageSoundC = target.DamageSound
	r.PlayerDamageSound = func(got, effective *Object) {
		if got != target || effective != weapon || target.Frame134 != 1403 {
			t.Fatal("sound lost live attribution/frame")
		}
		events = append(events, "sound")
		frame++
	}
	r.ShieldReduce = func(got *Object, damage *int32, typ object.DamageType, effective *Object) {
		if got != target || effective != weapon || typ != object.DamageFlame || *damage != 10 {
			t.Fatal("Shield reordered damage")
		}
		events = append(events, "shield")
		*damage -= 3
	}
	clear := r.DamageClear
	r.DamageClear = func(got *Object, damage int32) { events = append(events, "HP"); clear(got, damage) }
	if !DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageFlame, r) || target.HealthData.Cur != 193 ||
		!slices.Equal(events, []string{"resistance", "invisible-off", "late-defense", "sound", "shield", "HP"}) {
		t.Fatalf("FLAME HP=%d events=%v", target.HealthData.Cur, events)
	}
}

func TestDefaultDamageWorld4E0B30MissileFlameImmunityAndFriendly(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		immune, gameplay, enemy bool
		want                    uint16
	}{
		{"fire-immune", true, true, true, 200}, {"campaign-friendly", false, false, false, 200},
		{"campaign-enemy", false, false, true, 191}, {"arena-friendly missile", false, true, false, 191},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, weapon, _ := damageFlameFixture4E17B0(t, true, false, false)
			if tc.immune {
				target.ObjSubClass |= 0x400
			}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.GameplayFlag1 = func() bool { return tc.gameplay }
			r.IsEnemy = func(*Object, *Object) bool { return tc.enemy }
			calls := 0
			r.FireProtection = func(*Object) float64 { calls++; return 0 }
			DefaultDamageWorld4E0B30(target, source, weapon, 9, object.DamageFlame, r)
			wantCalls := 1
			if tc.want == 200 {
				wantCalls = 0
			}
			if target.HealthData.Cur != tc.want || calls != wantCalls ||
				!target.UpdateDataMonster().StatusFlags.Has(object.MonStatusOnFire) {
				t.Fatalf("HP=%d protection calls=%d on-fire=%x", target.HealthData.Cur, calls, target.UpdateDataMonster().StatusFlags)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MissileFlameRounding(t *testing.T) {
	for _, tc := range []struct {
		damage     int32
		resistance float64
		want       int32
	}{
		{9, 0.5, 4}, {11, 0.5, 6}, {1, 0.5, 1}, {0, 0, 1}, {-9, 0.5, -4}, {9, 1, 1}, {9, -0.5, 14},
	} {
		t.Run(fmt.Sprintf("%d/%g", tc.damage, tc.resistance), func(t *testing.T) {
			target, source, weapon, _ := damageFlameFixture4E17B0(t, true, false, false)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return tc.resistance }
			var got int32
			r.DamageClear = func(_ *Object, d int32) { got = d }
			DefaultDamageWorld4E0B30(target, source, weapon, tc.damage, object.DamageFlame, r)
			if got != tc.want {
				t.Fatalf("damage=%d want=%d", got, tc.want)
			}
		})
	}
}
