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

func playerDamageElectricSelfArmorFixture4E17B0(t *testing.T, playerTarget, playerSource bool) (target, source, armor *Object, carry *float32, modifier *ModifierEff) {
	t.Helper()
	target, source = defaultDamageElectricSelfFixture4E0B30(t, playerTarget, playerSource)
	if playerTarget {
		ud := target.UpdateDataPlayer()
		ud.Field21, ud.Field57 = math.Float32bits(0.25), math.Float32bits(0.4)
		ud.Field76, ud.Field75 = 99, 77
	} else {
		ud := target.UpdateDataMonster()
		ud.Field1, ud.Field518 = math.Float32bits(0.25), math.Float32bits(0.4)
	}
	carry = new(float32)
	*carry = 0.25
	modifier = &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor = &Object{
		ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(carry),
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, modifier, nil, nil}}),
	}
	target.InvFirstItem = armor
	return
}

func playerDamageElectricSelfMetadata4E17B0(target *Object) (marker, typ, carry uint32) {
	if target.Class().Has(object.ClassPlayer) {
		ud := target.UpdateDataPlayer()
		return ud.Field76, ud.Field75, ud.Field21
	}
	ud := target.UpdateDataMonster()
	return ud.Field547, ud.Field546, ud.Field1
}

func playerDamageElectricSelfArmorRuntime4E17B0(t *testing.T, target, source, armor *Object, modifier *ModifierEff, typ object.DamageType, events *[]string, damages *[]int32) PlayerDamageRuntime4E17B0 {
	t.Helper()
	r := playerDamageRuntime4E17B0(t, nil, damages)
	r.ElectricArmorScale = func(got *Object) float32 {
		if got != target {
			t.Fatal("electric scale target")
		}
		*events = append(*events, "scale")
		return 0.75
	}
	r.ItemArmorValue = func(got *Object) float32 {
		if got != armor {
			t.Fatal("armor definition target")
		}
		return 0.4
	}
	r.ApplyArmorDefend = func(m *ModifierEff, item, victim, weapon, attacker *Object, portion *float32) bool {
		if m != modifier || item != armor || victim != target || weapon != source || attacker != source || *portion != 5 {
			t.Fatal("electric armor defense must receive the raw damage and original self-weapon")
		}
		*events = append(*events, "armor-defend")
		*portion += 0.5
		return true
	}
	r.CanDamageArmor = func(got *Object) bool { return got == armor }
	r.DamageArmor = func(item, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
		marker, _, _ := playerDamageElectricSelfMetadata4E17B0(target)
		if item != armor || attacker != source || weapon != source || damage != 6 || gotType != typ || marker != 0 || *(*float32)(item.UpdateData) != -0.25 {
			t.Fatal("electric armor callback lost self-weapon, raw damage, carry or prefix order")
		}
		*events = append(*events, "armor-damage")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	r.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
		if owner != target || item != armor || before != 100 || after != 94 {
			t.Fatal("electric armor report")
		}
		*events = append(*events, "report")
	}
	return r
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricArmorAndDefault(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, playerSource := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("player-target-%t/player-source-%t/%s", playerTarget, playerSource, typ), func(t *testing.T) {
					target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, playerTarget, playerSource)
					var events []string
					var damages []int32
					r := playerDamageElectricSelfArmorRuntime4E17B0(t, target, source, armor, modifier, typ, &events, &damages)
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = func() float32 { events = append(events, "quest"); return 0.5 }
					r.DefaultDamage = func(victim, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
						marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
						if victim != target || attacker != source || weapon != source || damage != 2 || gotType != typ || marker != 2 || rawType != uint32(typ) || residual != 0 || armor.HealthData.Cur != 94 || target.HealthData.Cur != 60 {
							t.Fatal("electric Quest/default entry lost self-weapon or ordered metadata")
						}
						events = append(events, "default")
						return DefaultDamageWorld4E0B30(victim, attacker, weapon, damage, gotType, DefaultDamageWorldRuntime4E0B30{
							Frame: r.Frame, GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
							ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0.25 },
							MonsterHasHitSound: func(got *Object) bool {
								if got != source || playerSource {
									t.Fatal("player source used monster sound lookup")
								}
								return false
							},
							PlayerSetState: func(*Object, PlayerState) bool { return true },
							DamageClear:    r.DamageClear, Unsupported: r.Unsupported,
						})
					}
					if handled, result := PlayerDamageNative4E17B0(target, source, source, 5, typ, r); !handled || !result {
						t.Fatalf("electric self-weapon=%t/%t", handled, result)
					}
					want := []string{"scale", "armor-defend", "armor-damage", "report", "quest", "default", "protection"}
					if !slices.Equal(events, want) || !slices.Equal(damages, []int32{2}) || target.HealthData.Cur != 58 || *carry != -0.25 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) {
						t.Fatalf("events=%v damage=%v HP=%d armor-carry=%g source=%p", events, damages, target.HealthData.Cur, *carry, target.Obj130)
					}
					if playerTarget {
						if ud := target.UpdateDataPlayer(); ud.Field40_0 != 2 || ud.Field40_1 != 0xabcd {
							t.Fatal("electric low WORD overwrote adjacent state")
						}
					} else if ud := target.UpdateDataMonster(); ud.Field523_2 != 2 || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("NPC electric state/injury")
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricArmorBeforeGodMode(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			t.Run(fmt.Sprintf("player-source-%t/%s", playerSource, typ), func(t *testing.T) {
				target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, true, playerSource)
				var events []string
				var damages []int32
				r := playerDamageElectricSelfArmorRuntime4E17B0(t, target, source, armor, modifier, typ, &events, &damages)
				r.GodMode = func() bool { events = append(events, "god"); return true }
				r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
					t.Fatal("GodMode entered default damage")
					return false
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, source, 5, typ, r); !handled || !result {
					t.Fatalf("GodMode electric=%t/%t", handled, result)
				}
				marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
				if !slices.Equal(events, []string{"scale", "armor-defend", "armor-damage", "report", "god"}) || marker != 2 || rawType != uint32(typ) || residual != 0 || armor.HealthData.Cur != 94 || *carry != -0.25 || len(damages) != 0 || target.HealthData.Cur != 60 {
					t.Fatalf("events=%v marker=%d/%d HP=%d", events, marker, rawType, target.HealthData.Cur)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricReflectNPC(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			target, source := defaultDamageElectricSelfFixture4E0B30(t, false, true)
			target.Buffs = 1 << playerDamageReflectEnchant4E17B0
			calls := 0
			reason := ""
			r := PlayerDamageRuntime4E17B0{
				ElectricArmorScale: func(*Object) float32 { return 1 },
				DefaultDamage: func(victim, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
					if victim != target || attacker != source || weapon != source || damage != 8 || gotType != typ {
						t.Fatal("NPC Reflect default arguments")
					}
					calls++
					return true
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
			}
			handled, result := PlayerDamageNative4E17B0(target, source, source, 8, typ, r)
			marker, rawType, _ := playerDamageElectricSelfMetadata4E17B0(target)
			if typ == object.DamageElectric {
				// 004E199A only enters the non-missile reflection effect for 16/17.
				if !handled || !result || calls != 1 || reason != "" || marker != 2 || rawType != uint32(typ) {
					t.Fatalf("ordinary electric was reflected: %t/%t calls=%d reason=%q", handled, result, calls, reason)
				}
			} else if handled || result || calls != 0 || reason != "monster Reflect Shield" || marker != 99 || rawType != 77 {
				t.Fatalf("unported NPC airborne reflection mutated: %t/%t calls=%d reason=%q", handled, result, calls, reason)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricReflectPlayer(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			for _, front := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/%s/front-%t", playerSource, typ, front), func(t *testing.T) {
					target, source := defaultDamageElectricSelfFixture4E0B30(t, true, playerSource)
					target.Buffs = 1 << playerDamageReflectEnchant4E17B0
					source.PosVec.X = 37
					var damages []int32
					r := playerDamageRuntime4E17B0(t, nil, &damages)
					r.ElectricArmorScale = func(*Object) float32 { return 1 }
					reflected, calls := false, 0
					r.BlockDirection = func(got *Object, position types.Pointf) bool {
						if typ != object.DamageAirborneElectric || got != target || position != source.PosVec {
							t.Fatal("self-weapon Reflect facing predicate")
						}
						return front
					}
					r.Audio = func(id int, got *Object) {
						if id != 122 || got != target || reflected {
							t.Fatal("self-weapon reflection audio")
						}
						reflected = true
					}
					r.DefaultDamage = func(victim, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
						if victim != target || attacker != source || weapon != source || damage != 8 || gotType != typ {
							t.Fatal("self-weapon reflection default arguments")
						}
						calls++
						return true
					}
					wantReflect := front && typ == object.DamageAirborneElectric
					if handled, result := PlayerDamageNative4E17B0(target, source, source, 8, typ, r); !handled || result == wantReflect || reflected != wantReflect {
						t.Fatalf("self-weapon reflection=%t/%t audio=%t want=%t", handled, result, reflected, wantReflect)
					}
					marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
					if wantReflect {
						if calls != 0 || marker != 0 || residual != 0 {
							t.Fatal("reflected self-weapon entered electric carry/default")
						}
					} else if calls != 1 || marker != 2 || rawType != uint32(typ) || residual != 0 {
						t.Fatal("unreflected self-weapon lost electric attribution")
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricMissingService(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, playerSource := range []bool{false, true} {
			for _, missing := range []string{"scale", "default", "quest", "armor defend", "armor callback"} {
				t.Run(fmt.Sprintf("player-target-%t/player-source-%t/%s", playerTarget, playerSource, missing), func(t *testing.T) {
					target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, playerTarget, playerSource)
					var events []string
					var damages []int32
					r := playerDamageElectricSelfArmorRuntime4E17B0(t, target, source, armor, modifier, object.DamageElectric, &events, &damages)
					r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
						t.Fatal("unsupported self-weapon entered default damage")
						return false
					}
					reason := ""
					r.Unsupported = func(why string, _, _, weapon *Object, _ int32, _ object.DamageType) {
						if weapon != source {
							t.Fatal("unsupported report lost the original self-weapon")
						}
						reason = why
					}
					switch missing {
					case "scale":
						r.ElectricArmorScale = nil
					case "default":
						r.DefaultDamage = nil
					case "quest":
						r.QuestMode, r.QuestDamageScale = func() bool { return true }, nil
					case "armor defend":
						r.ApplyArmorDefend = nil
					case "armor callback":
						r.CanDamageArmor = nil
					}
					if handled, result := PlayerDamageNative4E17B0(target, source, source, 5, object.DamageElectric, r); handled || result || reason == "" {
						t.Fatalf("missing service=%t/%t reason=%q", handled, result, reason)
					}
					marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
					if marker != 99 || rawType != 77 || residual != math.Float32bits(0.25) || *carry != 0.25 || armor.HealthData.Cur != 100 || target.HealthData.Cur != 60 || target.Obj130 != nil || len(damages) != 0 {
						t.Fatal("unsupported self-weapon changed damage state")
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0SelfWeaponElectricAdmission(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, variant := range []string{"missing update", "distinct weapon", "weapon caster", "wand caster", "missile caster", "non-electric"} {
			t.Run(fmt.Sprintf("player-target-%t/%s", playerTarget, variant), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, playerTarget, true)
				weapon, typ := source, object.DamageElectric
				switch variant {
				case "missing update":
					source.UpdateData = nil
				case "distinct weapon":
					weapon = &Object{ObjClass: object.ClassPlayer}
				case "weapon caster":
					source.ObjClass |= object.ClassWeapon
				case "wand caster":
					source.ObjClass |= object.ClassWand
				case "missile caster":
					source.ObjClass |= object.ClassMissile
				case "non-electric":
					typ = object.DamagePlasma
				}
				beforeMarker, beforeType, beforeCarry := playerDamageElectricSelfMetadata4E17B0(target)
				reason := ""
				r := PlayerDamageRuntime4E17B0{
					ElectricArmorScale: func(*Object) float32 { t.Fatal("unported shape reached electric carry"); return 1 },
					DefaultDamage: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
						t.Fatal("unported shape entered default damage")
						return false
					},
					Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 8, typ, r); handled || result || reason == "" {
					t.Fatalf("unsupported shape=%t/%t reason=%q", handled, result, reason)
				}
				marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
				if marker != beforeMarker || rawType != beforeType || residual != beforeCarry || target.HealthData.Cur != 60 || target.Obj130 != nil {
					t.Fatal("unported self-weapon shape mutated target")
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0UnarmedMonsterElectricNPC(t *testing.T) {
	target, source := defaultDamageElectricSelfFixture4E0B30(t, false, false)
	var damages []int32
	r := playerDamageRuntime4E17B0(t, nil, &damages)
	r.ElectricArmorScale = func(*Object) float32 { return 1 }
	r.DefaultDamage = func(victim, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
		if victim != target || attacker != source || weapon != nil || damage != 8 || typ != object.DamageElectric {
			t.Fatal("unarmed monster electric NPC arguments")
		}
		return DefaultDamageWorld4E0B30(victim, attacker, weapon, damage, typ, DefaultDamageWorldRuntime4E0B30{
			GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
			ElectricProtection: func(*Object) float64 { return 0 }, DamageClear: r.DamageClear, Unsupported: r.Unsupported,
		})
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, nil, 8, object.DamageElectric, r); !handled || !result || target.HealthData.Cur != 52 || !slices.Equal(damages, []int32{8}) {
		t.Fatalf("unarmed monster electric NPC=%t/%t HP=%d damage=%v", handled, result, target.HealthData.Cur, damages)
	}
}
