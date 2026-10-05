package server

import (
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDamageWorld4E0B30PlayerWeaponlessExplosionMatrix(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, tc := range []struct {
			name       string
			damage     int32
			protection float64
			want       int32
		}{
			{"ordinary", 45, 0, 45}, {"tie-even", 9, 0.5, 4}, {"tie-odd", 11, 0.5, 6},
			{"zero", 0, 0, 1}, {"minimum", 1, 1, 1}, {"signed", -9, 0.5, -4},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, true)
				source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
				r := damageMeleeWorldRuntime4E0B30(t)
				r.FireProtection = func(*Object) float64 { return tc.protection }
				amount := int32(-12345)
				r.DamageClear = func(v *Object, d int32) {
					if v != target {
						t.Fatal("wrong HP target")
					}
					amount = d
				}
				if !DefaultDamageWorld4E0B30(target, source, nil, tc.damage, object.DamageExplosion, r) {
					t.Fatal("weapon-less player explosion returned false")
				}
				pos := types.Pointf{}
				if source != nil {
					pos = source.PrevPos
				}
				if amount != tc.want || target.Obj130 != source || target.Pos132 != pos || target.Field131 != 7 || target.Frame134 != 1400 {
					t.Fatalf("damage=%d want=%d source=%p/%p pos=%v/%v type/frame=%d/%d", amount, tc.want, target.Obj130, source, target.Pos132, pos, target.Field131, target.Frame134)
				}
				wantState := PlayerState13
				if tc.want >= 20 {
					wantState = PlayerState30
				}
				if target.UpdateDataPlayer().State != wantState {
					t.Fatal("incorrect player hurt state")
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerWeaponlessExplosionOwnerGate(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, gameplay := range []bool{false, true} {
			for _, quest := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/gameplay-%t/quest-%t", kind, gameplay, quest), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, true)
					source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
					r := damageMeleeWorldRuntime4E0B30(t)
					r.GameplayFlag1 = func() bool { return gameplay }
					r.QuestMode = func() bool { return quest }
					queries := 0
					r.IsEnemy = func(v, s *Object) bool {
						if v != target || s != source {
							t.Fatal("owner gate identity lost")
						}
						queries++
						return false
					}
					r.FireProtection = func(*Object) float64 { return 0 }
					DefaultDamageWorld4E0B30(target, source, nil, 9, object.DamageExplosion, r)
					wantHP := uint16(191)
					if !gameplay && source != nil && (source != target || quest) {
						wantHP = 200
					}
					wantQueries := 0
					if !gameplay && source != nil {
						wantQueries++
					}
					if kind == "npc" && wantHP < 200 {
						wantQueries++
					} // Live aggression latch only.
					if target.HealthData.Cur != wantHP || queries != wantQueries {
						t.Fatalf("HP=%d/%d enemy queries=%d/%d", target.HealthData.Cur, wantHP, queries, wantQueries)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerWeaponlessExplosionOrder(t *testing.T) {
	for _, sourced := range []bool{false, true} {
		t.Run(fmt.Sprintf("source-%t", sourced), func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			var source *Object
			if sourced {
				source = damageMeleeUnitFixture4E17B0(t, true)
			}
			modifier := &ModifierEff{}
			armor := damageMeleeArmorFixture4E17B0(target, 0, 0)
			modifier.Defend76.Fnc = armor.Damage
			armor.InitDataModifier().Modifiers[2] = modifier
			target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			frame := uint32(1400)
			r.Frame = func() uint32 { return frame }
			r.FireProtection = func(*Object) float64 { events = append(events, "fire"); return 0.5 }
			r.Audio = func(id int, v *Object) {
				if id != 104 || v != target {
					t.Fatal("unexpected protection sound")
				}
				events = append(events, "fire-sound")
			}
			r.BuffOff = func(v *Object, e EnchantID) {
				if v != target || e != defaultDamageInvisibleEnchant4E0B30 || target.Pos132 != source.PrevPos {
					t.Fatal("BuffOff precedes position")
				}
				events = append(events, "buff")
			}
			r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == modifier }
			r.ApplyLateDefend = func(m *ModifierEff, item, owner, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
				if m != modifier || item != armor || owner != target || weapon != nil || attacker != source || d != 22 || typ != object.DamageExplosion {
					t.Fatal("late defense identity/order lost")
				}
				events = append(events, "defend")
				frame = 1404
				return 24
			}
			r.DefaultDamageSound = func(v, s *Object) {
				if v != target || s != source || target.Obj130 != source || target.Frame134 != 1404 {
					t.Fatal("sound precedes live attribution")
				}
				events = append(events, "sound")
			}
			r.PlayerSetState = func(v *Object, state PlayerState) bool {
				if v != target || state != PlayerState30 {
					t.Fatal("wrong hurt state")
				}
				events = append(events, "hurt")
				return true
			}
			r.ShieldReduce = func(v *Object, d *int32, typ object.DamageType, s *Object) {
				if v != target || *d != 24 || typ != object.DamageExplosion || s != source {
					t.Fatal("wrong live Shield input")
				}
				events = append(events, "shield")
				*d = 0
			}
			r.DamageClear = func(*Object, int32) { t.Fatal("zero Shield result reached HP") }
			if DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r) || target.HealthData.Cur != 200 {
				t.Fatal("zero Shield must leave HP unchanged and return false")
			}
			want := []string{"fire", "fire-sound"}
			if sourced {
				want = append(want, "buff")
			}
			want = append(want, "defend", "sound", "hurt", "shield")
			if !slices.Equal(events, want) {
				t.Fatalf("events=%v want=%v", events, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerWeaponlessExplosionNoShock(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
	target.Buffs = 1 << defaultDamageShockEnchant4E0B30
	r := damageMeleeWorldRuntime4E0B30(t)
	r.FireProtection = func(*Object) float64 { return 0 }
	r.CallDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("weapon-less explosion retaliated with Shock")
		return true
	}
	if !DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r) || target.HealthData.Cur != 155 || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
		t.Fatal("weapon-less explosion lost HP or consumed Shock")
	}
}

func TestDefaultDamageWorld4E0B30PlayerWeaponlessExplosionMissingServices(t *testing.T) {
	for _, missing := range []string{"damage-clear", "fire", "buff", "hit-sound", "hurt"} {
		t.Run(missing, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 { return 0 }
			switch missing {
			case "damage-clear":
				r.DamageClear = nil
			case "fire":
				r.FireProtection = nil
			case "buff":
				r.BuffOff = nil
			case "hit-sound":
				r.MonsterHasHitSound = nil
			case "hurt":
				r.PlayerSetState = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			DefaultDamageWorld4E0B30(target, source, nil, 45, object.DamageExplosion, r)
			if reason == "" || target.HealthData.Cur != 200 {
				t.Fatal("missing service silently consumed a hit")
			}
			if missing != "hurt" && (target.Obj130 != nil || target.Frame134 != 0) {
				t.Fatal("admission failure partially attributed damage")
			}
		})
	}
}
