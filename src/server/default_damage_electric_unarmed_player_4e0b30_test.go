package server

import (
	"fmt"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// 004E0D46's Shock retaliation calls the attacker's damage callback with the
// enchanted unit as source and a nil weapon. Do not substitute a self-weapon:
// that changes late-defense arguments and can trigger a second Shock.
func TestDefaultDamageWorld4E0B30UnarmedPlayerElectricDefense(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			t.Run(fmt.Sprintf("player-target-%t/%s", playerTarget, typ), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, playerTarget, true)
				target.Buffs = 1<<defaultDamageShieldEnchant4E0B30 | 1<<defaultDamageShockEnchant4E0B30
				modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}})}
				target.InvFirstItem = armor
				var events []string
				r := DefaultDamageWorldRuntime4E0B30{
					Frame: func() uint32 { return 702 }, GameplayFlag1: func() bool { return true },
					IsEnemy: func(a, b *Object) bool { return a == target && b == source },
					ElectricProtection: func(got *Object) float64 {
						if got != target {
							t.Fatal("protection target")
						}
						events = append(events, "protection")
						return 0.25
					},
					BuffOff: func(got *Object, enchant EnchantID) {
						if got != target || enchant != 0 {
							t.Fatal("nil-weapon electric must only clear invisibility")
						}
						if playerTarget {
							if got.UpdateDataPlayer().Field40_0 != 2 || got.UpdateDataPlayer().Field40_1 != 0xabcd {
								t.Fatal("electric low-word store")
							}
						} else if got.UpdateDataMonster().Field523_2 != 2 || got.UpdateDataMonster().Field547 != 0 {
							t.Fatal("NPC electric prefix")
						}
						events = append(events, "buff-off")
					},
					CanApplyLateDefend: func(got *ModifierEff) bool { return got == modifier },
					ApplyLateDefend: func(m *ModifierEff, item, victim, weapon, attacker *Object, damage int32, gotType object.DamageType) int32 {
						if m != modifier || item != armor || victim != target || weapon != nil || attacker != source || damage != 24 || gotType != typ {
							t.Fatal("nil weapon was lost in late defense")
						}
						events = append(events, "late-defend")
						return damage + 2
					},
					MonsterHasHitSound: func(*Object) bool { t.Fatal("player source read as monster"); return false },
					DefaultDamageSound: func(victim, attacker *Object) {
						if victim != target || attacker != source || target.Obj130 != source || target.Field131 != uint32(typ) || target.Frame134 != 702 {
							t.Fatal("electric attribution/sound order")
						}
						events = append(events, "sound")
					},
					GameBallOnDamage: func(attacker, victim *Object, damage int32) {
						if !playerTarget || attacker != source || victim != target || damage != 26 {
							t.Fatal("GameBall arguments")
						}
						events = append(events, "ball")
					},
					PlayerSetState: func(victim *Object, state PlayerState) bool {
						if !playerTarget || victim != target || state != PlayerState30 {
							t.Fatal("player hurt arguments")
						}
						events = append(events, "hurt")
						return true
					},
					ShieldReduce: func(victim *Object, damage *int32, gotType object.DamageType, attacker *Object) {
						if victim != target || *damage != 26 || gotType != typ || attacker != source {
							t.Fatal("nil-weapon Shield source")
						}
						events = append(events, "shield")
						*damage = 5
					},
					DamageClear: func(victim *Object, damage int32) {
						if victim != target || damage != 5 {
							t.Fatal("electric damage tail")
						}
						events = append(events, "damage")
						victim.HealthData.Cur -= uint16(damage)
					},
					Audio: func(int, *Object) { t.Fatal("nil weapon caused Shock or off-frame protection sound") },
					CallDamage: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
						t.Fatal("nil weapon caused recursive Shock")
						return false
					},
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Fatalf("nil-weapon player electric rejected: %s", reason)
					},
				}
				if !DefaultDamageWorld4E0B30(target, source, nil, 32, typ, r) {
					t.Fatal("electric result")
				}
				want := []string{"protection", "buff-off", "late-defend", "sound"}
				if playerTarget {
					want = append(want, "ball", "hurt")
				}
				want = append(want, "shield", "damage")
				if !slices.Equal(events, want) || target.HealthData.Cur != 55 || target.Pos132 != source.PrevPos || !target.HasEnchant(defaultDamageShockEnchant4E0B30) {
					t.Fatalf("events=%v HP=%d position=%v Shock=%t", events, target.HealthData.Cur, target.Pos132, target.HasEnchant(defaultDamageShockEnchant4E0B30))
				}
				if !playerTarget {
					ud := target.UpdateDataMonster()
					if ud.Field547 != 2 || ud.Field546 != uint32(typ) || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("electric raw-type injury latch")
					}
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30UnarmedPlayerElectricFriendly(t *testing.T) {
	for _, gameplay := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			t.Run(fmt.Sprintf("gameplay-%t/%s", gameplay, typ), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, true, true)
				calls := 0
				r := DefaultDamageWorldRuntime4E0B30{
					GameplayFlag1: func() bool { return gameplay }, IsEnemy: func(*Object, *Object) bool { return false },
					ElectricProtection: func(*Object) float64 { return 0 },
					PlayerSetState:     func(*Object, PlayerState) bool { t.Fatal("small electric hit hurt state"); return false },
					DamageClear:        func(victim *Object, damage int32) { calls++; victim.HealthData.Cur -= uint16(damage) },
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Fatalf("nil-weapon electric rejected: %s", reason)
					},
				}
				if !DefaultDamageWorld4E0B30(target, source, nil, 8, typ, r) {
					t.Fatal("friendly result")
				}
				want := 0
				if gameplay {
					want = 1
				}
				if calls != want || target.HealthData.Cur != 60-uint16(want*8) {
					t.Fatalf("calls=%d HP=%d want-hit=%d", calls, target.HealthData.Cur, want)
				}
			})
		}
	}
}
