package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultDamageElectricSelfFixture4E0B30(t *testing.T, playerTarget, playerSource bool) (*Object, *Object) {
	t.Helper()
	target, source, _ := playerDamageFixture4E17B0(t)
	target.DamageSound = nil
	target.HealthData.Cur, target.HealthData.Max = 60, 60
	if playerTarget {
		update := target.UpdateDataPlayer()
		update.Field40_0, update.Field40_1 = 0x1122, 0xabcd
		update.Field76, update.Field75 = 2, 9
	} else {
		target.ObjClass, target.ObjSubClass = object.ClassMonster, 0x11012
		target.UpdateData = unsafe.Pointer(&MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 99})
	}
	if playerSource {
		source = &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{}})}
	}
	source.TypeInd, source.PrevPos = 713, types.Pointf{X: 12.5, Y: -40.25}
	return target, source
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricDefense(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, playerSource := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("player-target-%t/player-source-%t/%s", playerTarget, playerSource, typ), func(t *testing.T) {
					target, source := defaultDamageElectricSelfFixture4E0B30(t, playerTarget, playerSource)
					target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
					modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
					armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}})}
					target.InvFirstItem = armor
					var events []string
					r := DefaultDamageWorldRuntime4E0B30{
						Frame: func() uint32 { return 702 }, GameplayFlag1: func() bool { return true },
						IsEnemy: func(a, b *Object) bool { return a == target && b == source },
						ElectricProtection: func(got *Object) float64 {
							if got != target {
								t.Fatal("electric protection target")
							}
							events = append(events, "protection")
							return 0.25
						},
						BuffOff: func(got *Object, enchant EnchantID) {
							if got != target || enchant != 0 {
								t.Fatal("invisibility arguments")
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
							if m != modifier || item != armor || victim != target || weapon != source || attacker != source || damage != 6 || gotType != typ {
								t.Fatal("electric late defense lost the self-weapon")
							}
							events = append(events, "late-defend")
							return damage + 2
						},
						MonsterHasHitSound: func(got *Object) bool {
							if got != source || playerSource {
								t.Fatal("non-monster sound lookup")
							}
							events = append(events, "hit-sound")
							return false
						},
						DefaultDamageSound: func(victim, attacker *Object) {
							if victim != target || attacker != source || target.Obj130 != source || target.Field131 != uint32(typ) || target.Frame134 != 702 {
								t.Fatal("electric attribution/sound order")
							}
							events = append(events, "sound")
						},
						PlayerSetState: func(*Object, PlayerState) bool { t.Fatal("small electric hit hurt state"); return false },
						ShieldReduce: func(victim *Object, damage *int32, gotType object.DamageType, attacker *Object) {
							if victim != target || *damage != 8 || gotType != typ || attacker != source {
								t.Fatal("electric Shield lost the self-weapon")
							}
							if !playerSource && source.UpdateDataMonster().Field130 != 702 {
								t.Fatal("monster attack timestamp")
							}
							events = append(events, "shield")
							*damage = 5
						},
						DamageClear: func(got *Object, damage int32) {
							if got != target || damage != 5 {
								t.Fatal("electric damage tail")
							}
							events = append(events, "damage")
							got.HealthData.Cur -= uint16(damage)
						},
						Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
							t.Fatalf("self-weapon electric rejected: %s", reason)
						},
					}
					if !DefaultDamageWorld4E0B30(target, source, source, 8, typ, r) {
						t.Fatal("self-weapon electric returned false")
					}
					want := []string{"protection", "buff-off", "late-defend"}
					if !playerSource {
						want = append(want, "hit-sound")
					}
					want = append(want, "sound", "shield", "damage")
					if !slices.Equal(events, want) || target.HealthData.Cur != 55 || target.Pos132 != source.PrevPos {
						t.Fatalf("events=%v HP=%d position=%v", events, target.HealthData.Cur, target.Pos132)
					}
					if !playerTarget {
						ud := target.UpdateDataMonster()
						if ud.Field547 != 2 || ud.Field546 != uint32(typ) || !ud.StatusFlags.Has(object.MonStatusInjured) {
							t.Fatal("self-weapon electric must record the raw type, not the caster's type")
						}
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricFriendly(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, playerSource := range []bool{false, true} {
			for _, enemy := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-target-%t/player-source-%t/enemy-%t", playerTarget, playerSource, enemy), func(t *testing.T) {
					target, source := defaultDamageElectricSelfFixture4E0B30(t, playerTarget, playerSource)
					calls := 0
					r := DefaultDamageWorldRuntime4E0B30{
						GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return enemy },
						ElectricProtection: func(*Object) float64 { return 0 },
						MonsterHasHitSound: func(*Object) bool { return false }, PlayerSetState: func(*Object, PlayerState) bool { return true },
						DamageClear: func(got *Object, damage int32) { calls++; got.HealthData.Cur -= uint16(damage) },
						Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
							t.Fatalf("friendly self-weapon rejected: %s", reason)
						},
					}
					if !DefaultDamageWorld4E0B30(target, source, source, 8, object.DamageElectric, r) {
						t.Fatal("friendly result")
					}
					want := 0
					if playerSource || enemy {
						want = 1
					}
					if calls != want || target.HealthData.Cur != 60-uint16(want*8) {
						t.Fatalf("calls=%d HP=%d want-hit=%d", calls, target.HealthData.Cur, want)
					}
				})
			}
		}
	}
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricShockAndImmunity(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		for _, immune := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-source-%t/immune-%t", playerSource, immune), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, false, playerSource)
				target.Buffs = 1 << defaultDamageShockEnchant4E0B30
				if immune {
					target.ObjSubClass |= 0x800
				}
				var events []string
				r := DefaultDamageWorldRuntime4E0B30{
					GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return true },
					ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0 },
					Audio: func(id int, got *Object) {
						if id != 135 || got != source {
							t.Fatal("Shock sound")
						}
						events = append(events, "shock-sound")
					},
					BuffOff: func(got *Object, enchant EnchantID) {
						if got != target {
							t.Fatal("Shock buff target")
						}
						if enchant == 22 {
							got.Buffs &^= 1 << 22
							events = append(events, "shock-off")
						} else if enchant == 0 {
							events = append(events, "invisible-off")
						} else {
							t.Fatal("Shock buff id")
						}
					},
					BalanceFloatInd: func(name string, index int) float64 {
						if name != "ShockDamage" || index != 4 {
							t.Fatal("Shock balance")
						}
						return 40.5
					},
					CallDamage: func(victim, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
						if victim != source || attacker != target || weapon != nil || damage != 40 || typ != object.DamageElectric {
							t.Fatal("Shock retaliation contract")
						}
						events = append(events, "retaliate")
						return true
					},
					DamageClear: func(got *Object, damage int32) {
						if got != target || damage != 8 {
							t.Fatal("Shock incoming HP")
						}
						events = append(events, "damage")
						got.HealthData.Cur -= uint16(damage)
					},
					Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
						t.Fatalf("Shock self-weapon rejected: %s", reason)
					},
				}
				if !DefaultDamageWorld4E0B30(target, source, source, 8, object.DamageElectric, r) {
					t.Fatal("Shock incoming result")
				}
				var want []string
				if !playerSource {
					want = []string{"shock-sound", "shock-off", "retaliate"}
				}
				if !immune {
					want = append(want, "protection", "invisible-off", "damage")
				}
				wantHP := uint16(52)
				if immune {
					wantHP = 60
				}
				if !slices.Equal(events, want) || target.HealthData.Cur != wantHP || target.HasEnchant(22) != playerSource {
					t.Fatalf("events=%v want=%v HP=%d Shock=%t", events, want, target.HealthData.Cur, target.HasEnchant(22))
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricCampaignGate(t *testing.T) {
	target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
	r := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return false }, IsEnemy: func(*Object, *Object) bool { return false },
		ElectricProtection: func(*Object) float64 { t.Fatal("campaign friendly hit passed gate"); return 0 },
		DamageClear:        func(*Object, int32) { t.Fatal("campaign friendly hit reached HP") },
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("campaign gate rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, source, 8, object.DamageElectric, r) || target.HealthData.Cur != 60 || target.Obj130 != nil {
		t.Fatal("campaign friendly hit mutation")
	}
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricAdmission(t *testing.T) {
	for _, variant := range []string{"missing source update", "distinct unit weapon", "weapon-class caster", "non-electric type"} {
		t.Run(variant, func(t *testing.T) {
			target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
			weapon, typ := source, object.DamageElectric
			switch variant {
			case "missing source update":
				source.UpdateData = nil
			case "distinct unit weapon":
				weapon = &Object{ObjClass: object.ClassPlayer}
			case "weapon-class caster":
				source.ObjClass |= object.ClassWeapon
			case "non-electric type":
				typ = object.DamagePlasma
			}
			reason := ""
			r := DefaultDamageWorldRuntime4E0B30{GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return true },
				DamageClear: func(*Object, int32) { t.Fatal("unported electric shape reached HP") },
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 8, typ, r) || reason == "" || target.HealthData.Cur != 60 || target.Obj130 != nil {
				t.Fatalf("reason=%q HP=%d source=%p", reason, target.HealthData.Cur, target.Obj130)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30SelfWeaponElectricMissingService(t *testing.T) {
	target, source := defaultDamageElectricSelfFixture4E0B30(t, true, true)
	target.UpdateDataPlayer().Field21 = math.Float32bits(0.25)
	reason := ""
	r := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return true },
		PlayerSetState: func(*Object, PlayerState) bool { return true },
		Unsupported:    func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
	}
	if !DefaultDamageWorld4E0B30(target, source, source, 8, object.DamageElectric, r) || reason != "missing electric-protection service" || target.UpdateDataPlayer().Field40_0 != 0x1122 || target.HealthData.Cur != 60 || target.Obj130 != nil {
		t.Fatalf("reason=%q electric=%d HP=%d", reason, target.UpdateDataPlayer().Field40_0, target.HealthData.Cur)
	}
}
