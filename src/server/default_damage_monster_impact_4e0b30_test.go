package server

import (
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

func monsterImpactTarget4E17B0(t *testing.T, kind string) *Object {
	t.Helper()
	u := damageMeleeUnitFixture4E17B0(t, kind == "player")
	if kind == "monster" {
		u.ObjSubClass = 0
	}
	return u
}

func TestDefaultDamageMonsterSelfImpact4E0B30(t *testing.T) {
	for _, kind := range []string{"player", "NPC", "monster"} {
		for _, raw := range []int32{33, 0, -3} {
			for _, hitSound := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/raw-%d/hit-sound-%t", kind, raw, hitSound), func(t *testing.T) {
					v, a := monsterImpactTarget4E17B0(t, kind), monsterImpactTarget4E17B0(t, "monster")
					r := damageMeleeWorldRuntime4E0B30(t)
					var events []string
					clear := r.DamageClear
					r.IsEnemy = func(target, source *Object) bool {
						if target != v || source != a {
							t.Fatal("self-weapon changed enemy arguments")
						}
						events = append(events, "enemy")
						return true
					}
					if kind != "player" {
						v.UpdateDataMonster().Field547 = 99
					}
					r.BuffOff = func(target *Object, enchant EnchantID) {
						if target != v || enchant != 0 || v.Pos132 != a.PrevPos {
							t.Fatal("position/visibility order")
						}
						events = append(events, "buff")
					}
					r.MonsterHasHitSound = func(source *Object) bool {
						if source != a || v.Obj130 != a || v.Field131 != 11 || v.Frame134 != 1400 {
							t.Fatal("sound/attribution order")
						}
						events = append(events, "lookup")
						return hitSound
					}
					r.DefaultDamageSound = func(target, weapon *Object) {
						if target != v || weapon != a {
							t.Fatal("self-weapon sound identity")
						}
						events = append(events, "sound")
					}
					r.GameBallOnDamage = func(source, target *Object, amount int32) {
						if source != a || target != v || amount != raw {
							t.Fatal("GameBall tail arguments")
						}
						events = append(events, "ball")
					}
					r.PlayerSetState = func(target *Object, state PlayerState) bool {
						if target != v || state != PlayerState30 {
							t.Fatal("hurt-state tail")
						}
						events = append(events, "hurt")
						return true
					}
					// Zero/signed tests record the original HP service argument;
					// only positive damage uses this fixture's HP subtraction.
					r.DamageClear = func(target *Object, amount int32) {
						if target != v || amount != raw || a.UpdateDataMonster().Field130 != 1400 {
							t.Fatal("damage/source combat latch order")
						}
						if amount > 0 {
							clear(target, amount)
						}
						events = append(events, "hp")
					}
					if !DefaultDamageWorld4E0B30(v, a, a, raw, object.DamageImpact, r) {
						t.Fatal("unshielded self-weapon IMPACT returned false")
					}
					want := []string{"enemy", "buff", "lookup"}
					if !hitSound {
						want = append(want, "sound")
					}
					if kind == "player" {
						want = append(want, "ball")
						if raw >= 20 {
							want = append(want, "hurt")
						}
					} else {
						ud := v.UpdateDataMonster()
						if ud.Field547 != 2 || ud.Field546 != 11 || !ud.StatusFlags.Has(object.MonStatusInjured) {
							t.Fatal("self-weapon must leave type marker, not a distinct weapon ID")
						}
					}
					want = append(want, "enemy", "hp")
					if !slices.Equal(events, want) || (raw > 0 && v.HealthData.Cur != 167) {
						t.Fatalf("events=%v want=%v HP=%d", events, want, v.HealthData.Cur)
					}
				})
			}
		}
	}
}

func TestDefaultDamageMonsterSelfImpactFriendlyGates4E0B30(t *testing.T) {
	for _, kind := range []string{"player", "NPC", "monster"} {
		for _, owned := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/player-owned-%t", kind, owned), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, kind), monsterImpactTarget4E17B0(t, "monster")
				r := damageMeleeWorldRuntime4E0B30(t)
				if owned {
					a.ObjOwner = damageMeleeUnitFixture4E17B0(t, true)
					r.GameplayFlag1 = func() bool { return false }
				}
				r.IsEnemy = func(*Object, *Object) bool { return false }
				r.BuffOff = func(*Object, EnchantID) { t.Fatal("friendly hit reached BuffOff") }
				r.DamageClear = func(*Object, int32) { t.Fatal("friendly hit damaged HP") }
				if !DefaultDamageWorld4E0B30(v, a, a, 40, object.DamageImpact, r) || v.HealthData.Cur != 200 || v.Obj130 != nil || a.UpdateDataMonster().Field130 != 0 {
					t.Fatal("friendly IMPACT mutated hit metadata")
				}
			})
		}
	}
}

func TestDefaultDamageMonsterSelfImpactShockShield4E0B30(t *testing.T) {
	for _, kind := range []string{"player", "NPC", "monster"} {
		for _, absorbed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/absorbed-%t", kind, absorbed), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, kind), monsterImpactTarget4E17B0(t, "monster")
				v.Buffs = 1<<defaultDamageShockEnchant4E0B30 | 1<<defaultDamageShieldEnchant4E0B30
				r := damageMeleeWorldRuntime4E0B30(t)
				shock, shield, hp := 0, 0, 0
				r.Audio = func(int, *Object) {}
				r.BalanceFloatInd = func(key string, index int) float64 {
					if key != "ShockDamage" || index != 4 {
						t.Fatal("wrong Shock balance")
					}
					return 9
				}
				r.BuffOff = func(target *Object, enchant EnchantID) { target.Buffs &^= 1 << enchant }
				r.CallDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
					if target != a || source != v || weapon != nil || amount != 9 || typ != object.DamageElectric {
						t.Fatal("self-weapon must qualify Shock retaliation")
					}
					shock++
					return true
				}
				r.ShieldReduce = func(target *Object, amount *int32, typ object.DamageType, weapon *Object) {
					if target != v || weapon != a || typ != object.DamageImpact || *amount != 40 || shock != 1 || a.UpdateDataMonster().Field130 != 1400 {
						t.Fatal("Shield lost self-weapon or post-Shock order")
					}
					shield++
					if absorbed {
						*amount = 0
					}
				}
				r.DamageClear = func(*Object, int32) { hp++ }
				result := DefaultDamageWorld4E0B30(v, a, a, 40, object.DamageImpact, r)
				wantHP := 1
				if absorbed {
					wantHP = 0
				}
				if result == absorbed || shock != 1 || shield != 1 || hp != wantHP || v.HasEnchant(defaultDamageShockEnchant4E0B30) {
					t.Fatalf("result=%t shock=%d shield=%d hp=%d", result, shock, shield, hp)
				}
			})
		}
	}
}
