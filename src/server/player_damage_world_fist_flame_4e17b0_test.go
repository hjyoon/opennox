package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func worldFistFlameFixture4E17B0(t *testing.T, player, flame bool) (*Object, *Object, *Object, int32, object.DamageType) {
	t.Helper()
	v, _ := selfArrowFixture4E17B0(t, player)
	v.HealthData.Cur, v.HealthData.Max, v.HealthData.Field2 = 2000, 2000, 2000
	a := &Object{TypeInd: 734, ObjClass: object.Class(4194816), ObjFlags: object.Flags(525)}
	w := &Object{TypeInd: 1217, ObjClass: object.Class(1048576), ObjFlags: object.Flags(25166348), PrevPos: types.Ptf(17, 9), PosVec: types.Ptf(19, 11)}
	damage, typ := int32(200), object.DamageCrush
	if flame {
		a.TypeInd, a.ObjClass, a.ObjFlags = 1399, 0, object.Flags(16777796)
		w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 695, object.Class(2621441), 1, object.Flags(553665028)
		damage, typ = 64, object.DamageFlame
		if player {
			v.ObjFlags = object.Flags(150995460)
		}
	}
	return v, a, w, damage, typ
}

func TestPlayerDamageWorldFistFlame4E17B0HPArmorAndProtection(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, flame := range []bool{false, true} {
			for _, absorption := range []float32{0, 0.5} {
				for _, protection := range []float64{0, 0.25, 1} {
					if !flame && protection != 0 {
						continue
					}
					for _, friendly := range []bool{false, true} {
						t.Run(fmt.Sprintf("player-%t/flame-%t/armor-%g/protect-%g/friendly-%t", player, flame, absorption, protection, friendly), func(t *testing.T) {
							v, a, w, damage, typ := worldFistFlameFixture4E17B0(t, player, flame)
							r := damageMeleeRuntimeFixture4E17B0(t)
							var armor *Object
							if absorption != 0 {
								armor = damageMeleeArmorFixture4E17B0(v, absorption, 0.25)
								armor.HealthData.Cur, armor.HealthData.Max = 2000, 2000
								damageMeleeArmorRuntime4E17B0(&r, armor, absorption)
							} else if player {
								v.UpdateDataPlayer().Field21 = math.Float32bits(0.25)
							} else {
								v.UpdateDataMonster().Field1 = math.Float32bits(0.25)
							}
							world := damageMeleeWorldRuntime4E0B30(t)
							world.FireProtection = func(*Object) float64 { return protection }
							world.IsEnemy = func(target, source *Object) bool {
								if target != v || source != a {
									t.Fatal("world damage source identity")
								}
								return !friendly
							}
							r.BlockSourceExcluded = func(o *Object) bool { return !flame && o == w }
							r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
								return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
							}
							beforeA, beforeW := *a, *w
							effective, wear := int32(200), int32(0)
							if absorption != 0 {
								effective, wear = 150, 50
							}
							if flame {
								effective, wear = int32(64*(1-protection)), 64
								if effective == 0 {
									effective = 1
								}
							}
							for hit := range 3 {
								h, result := PlayerDamageNative4E17B0(v, a, w, damage, typ, r)
								marker, kind, carry := damageMeleeMarker4E17B0(v)
								if !h || !result || v.HealthData.Cur != uint16(2000-int32(hit+1)*effective) || marker != 1 || kind != uint32(w.TypeInd) || carry != 0.25 || v.Obj130 != w || v.Field131 != uint32(typ) || v.Pos132 != w.PrevPos || *a != beforeA || *w != beforeW {
									t.Fatalf("world hit=%d handled/result=%t/%t HP=%d marker=%d/%d carry=%g", hit, h, result, v.HealthData.Cur, marker, kind, carry)
								}
							}
							if armor != nil && armor.HealthData.Cur != uint16(2000-3*wear) {
								t.Fatalf("armor HP=%d want=%d", armor.HealthData.Cur, 2000-3*wear)
							}
						})
					}
				}
			}
		}
	}
}

func TestPlayerDamageImaginaryFlame4E17B0NPCImmunity(t *testing.T) {
	v, a, w, damage, typ := worldFistFlameFixture4E17B0(t, false, true)
	v.ObjSubClass |= 0x400
	armor := damageMeleeArmorFixture4E17B0(v, 0.5, 0.25)
	armor.HealthData.Cur, armor.HealthData.Max = 2000, 2000
	r := damageMeleeRuntimeFixture4E17B0(t)
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
	world := damageMeleeWorldRuntime4E0B30(t)
	world.FireProtection = func(*Object) float64 { t.Fatal("fire immunity must precede protection"); return 0 }
	r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
	}
	for range 3 {
		h, result := PlayerDamageNative4E17B0(v, a, w, damage, typ, r)
		marker, kind, carry := damageMeleeMarker4E17B0(v)
		// DefaultDamage resets the NPC latch before its fire-immunity exit;
		// the earlier PlayerDamage weapon type and armor wear remain visible.
		if !h || !result || v.HealthData.Cur != 2000 || marker != 0 || kind != 695 || carry != 0.25 || v.Obj130 != nil {
			t.Fatal("immune NPC prefix/tail ordering changed")
		}
	}
	if armor.HealthData.Cur != 1808 {
		t.Fatalf("fire immune NPC armor HP=%d", armor.HealthData.Cur)
	}
}

func TestPlayerDamageWorldFistFlameShape4E17B0Boundaries(t *testing.T) {
	for _, tc := range []struct {
		name           string
		source, weapon object.Class
		typ            object.DamageType
		self, want     bool
	}{
		{"pressure fist", object.Class(4194816), object.ClassSimple, object.DamageCrush, false, true},
		{"simple world fist", object.ClassSimple, object.ClassSimple, object.DamageCrush, false, true},
		{"imaginary flame", 0, object.Class(2621441), object.DamageFlame, false, true},
		{"unit fist", object.ClassPlayer, object.ClassSimple, object.DamageCrush, false, false},
		{"world unit weapon", object.ClassImmobile, object.ClassSimple | object.ClassMonster, object.DamageCrush, false, false},
		{"world weapon", object.ClassImmobile, object.ClassSimple | object.ClassWeapon, object.DamageCrush, false, false},
		{"world wand", object.ClassImmobile, object.ClassSimple | object.ClassWand, object.DamageCrush, false, false},
		{"self simple", object.ClassSimple, object.ClassSimple, object.DamageCrush, true, false},
		{"imaginary melee flame", 0, object.ClassWeapon, object.DamageFlame, false, false},
		{"imaginary wand missile", 0, object.ClassMissile | object.ClassWand, object.DamageFlame, false, false},
		{"imaginary unit missile", 0, object.ClassMissile | object.ClassMonster, object.DamageFlame, false, false},
		{"other type", 0, object.ClassMissile, object.DamageBlade, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, w := &Object{TypeInd: 9998, ObjClass: tc.source}, &Object{TypeInd: 9999, ObjClass: tc.weapon}
			if tc.self {
				w = a
			}
			if got := playerDamageWorldProjectileShape4E17B0(a, w, tc.typ); got != tc.want {
				t.Fatalf("admitted=%t want=%t", got, tc.want)
			}
		})
	}
}
