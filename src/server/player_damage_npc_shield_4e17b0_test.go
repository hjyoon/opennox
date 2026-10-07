package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func npcShieldFixture4E17B0(t *testing.T) (target, source, missile, shield *Object) {
	t.Helper()
	target, source = defaultDamageElectricSelfFixture4E0B30(t, false, true)
	ud := target.UpdateDataMonster()
	ud.ArmorEquipFlags, ud.AIStackInd = 0x1000000, 1
	ud.AIStack[0].Action, ud.AIStack[1].Action = uint32(ai.ACTION_GUARD), uint32(ai.ACTION_BLOCK_ATTACK)
	ud.Field1 = math.Float32bits(0.25)
	shield = &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 20, Max: 20}}
	later := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped}
	shield.InvNextItem = later
	target.InvFirstItem = &Object{ObjSubClass: 2, InvNextItem: shield}
	missile = &Object{TypeInd: 555, ObjClass: object.ClassMissile, PrevPos: types.Pointf{X: 20, Y: 3}, PosVec: types.Pointf{X: -20}}
	return
}

func npcShieldRuntime4E17B0(t *testing.T, target, source, attack, shield *Object, amount int32, typ object.DamageType, broken bool, events *[]string) PlayerDamageRuntime4E17B0 {
	t.Helper()
	ud := target.UpdateDataMonster()
	return PlayerDamageRuntime4E17B0{
		BlockSourceExcluded: func(got *Object) bool {
			if got != attack {
				t.Fatal("NPC shield exclusion used the wrong attack")
			}
			return false
		},
		BlockDirection: func(got *Object, pos types.Pointf) bool {
			if got != target || pos != attack.PrevPos {
				t.Fatal("NPC shield must face attack PrevPos, not current Pos")
			}
			return true
		},
		CanDamageBlockItem: func(got *Object) bool {
			if got != shield || ud.Field547 != 99 {
				t.Fatal("shield admission changed the hit prefix")
			}
			return true
		},
		Audio: func(id int, got *Object) {
			if id != 878 || got != target || ud.Field547 != 1 || ud.Field546 != uint32(attack.TypeInd) {
				t.Fatal("NPC shield audio/marker order")
			}
			*events = append(*events, "audio")
		},
		ProjectileReflect: func(got, reflector *Object) {
			if got != attack || reflector != target {
				t.Fatal("NPC shield reflection identity")
			}
			*events = append(*events, "reflect")
		},
		ClearOwner: func(got *Object) {
			if got != attack {
				t.Fatal("NPC shield clear owner identity")
			}
			*events = append(*events, "clear")
		},
		SetOwner: func(owner, got *Object) {
			if owner != target || got != attack {
				t.Fatal("NPC shield owner argument order")
			}
			*events = append(*events, "set")
		},
		BlockDamagePercent: func() float64 { *events = append(*events, "percent"); return 0.2 },
		DamageBlockItem: func(item, owner, attacker, effective *Object, wear float32, gotType object.DamageType) bool {
			if item != shield || owner != target || attacker != source || effective != attack || wear != float32(0.2*float64(amount)) || gotType != typ {
				t.Fatal("NPC shield durability arguments")
			}
			*events = append(*events, "durability")
			if broken {
				item.ObjFlags |= object.FlagDestroyed
			}
			return true
		},
		Melee: PlayerDamageMeleeRuntime4E17B0{MonsterPopBlockAction: func(got *Object) {
			if got != target || !shield.Flags().Has(object.FlagDestroyed) || ud.AIStackInd != 1 {
				t.Fatal("NPC shield pop preceded break")
			}
			*events = append(*events, "pop")
			ud.AIStackInd--
		}},
		PlayerSetState: func(*Object, PlayerState) bool { t.Fatal("NPC used player update layout"); return false },
		ObserveClear:   func(*Object) { t.Fatal("NPC used player observer") },
		DefaultDamage: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
			t.Fatal("blocked NPC hit reached HP damage")
			return false
		},
		Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
	}
}

func TestPlayerDamageNative4E17B0NPCShieldMissileMatrix(t *testing.T) {
	for _, armor := range []uint32{0x1000000, 0x2000000} {
		for _, flags := range []uint32{0, 2, 0x10, 0x20, 0x40, 0x70, 0x80} {
			for _, damage := range []int32{0, 9, -5} {
				for _, broken := range []bool{false, true} {
					for _, typ := range []object.DamageType{object.DamageImpact, object.DamageImpale, object.DamageZapRay} {
						t.Run(fmt.Sprintf("armor-%x/missile-%x/damage-%d/broken-%t/%s", armor, flags, damage, broken, typ), func(t *testing.T) {
							target, source, attack, shield := npcShieldFixture4E17B0(t)
							ud := target.UpdateDataMonster()
							ud.ArmorEquipFlags, attack.ObjSubClass = armor, object.SubClass(flags)
							before := *ud
							var events []string
							r := npcShieldRuntime4E17B0(t, target, source, attack, shield, damage, typ, broken, &events)
							if handled, result := PlayerDamageNative4E17B0(target, source, attack, damage, typ, r); !handled || result {
								t.Fatalf("NPC shield=%t/%t", handled, result)
							}
							want := []string{"audio"}
							if flags&0x70 == 0 {
								want = append(want, "reflect")
								if flags&2 == 0 {
									want = append(want, "clear", "set")
								}
							}
							want = append(want, "percent", "durability")
							before.Field547, before.Field546 = 1, uint32(attack.TypeInd)
							if broken {
								want = append(want, "pop")
								before.AIStackInd--
							}
							if !slices.Equal(events, want) || *ud != before || target.HealthData.Cur != 60 {
								t.Fatalf("events=%v want=%v HP=%d", events, want, target.HealthData.Cur)
							}
						})
					}
				}
			}
		}
	}
}

func TestPlayerDamageNative4E17B0NPCShieldMissingServiceDoesNotMutate(t *testing.T) {
	for _, service := range []string{"excluded", "direction", "reflect", "clear", "set", "audio", "percent", "can-wear", "wear", "pop"} {
		t.Run(service, func(t *testing.T) {
			target, source, attack, shield := npcShieldFixture4E17B0(t)
			before, itemBefore, attackBefore := *target.UpdateDataMonster(), *shield, *attack
			var events []string
			r := npcShieldRuntime4E17B0(t, target, source, attack, shield, 9, object.DamageImpact, false, &events)
			switch service {
			case "excluded":
				r.BlockSourceExcluded = nil
			case "direction":
				r.BlockDirection = nil
			case "reflect":
				r.ProjectileReflect = nil
			case "clear":
				r.ClearOwner = nil
			case "set":
				r.SetOwner = nil
			case "audio":
				r.Audio = nil
			case "percent":
				r.BlockDamagePercent = nil
			case "can-wear":
				r.CanDamageBlockItem = nil
			case "wear":
				r.DamageBlockItem = nil
			case "pop":
				r.Melee.MonsterPopBlockAction = nil
			}
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			if handled, result := PlayerDamageNative4E17B0(target, source, attack, 9, object.DamageImpact, r); handled || result || reason == "" {
				t.Fatalf("missing %s=%t/%t/%q", service, handled, result, reason)
			}
			if len(events) != 0 || *target.UpdateDataMonster() != before || *shield != itemBefore || *attack != attackBefore || target.HealthData.Cur != 60 {
				t.Fatalf("missing %s mutated state", service)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCShieldAdmission(t *testing.T) {
	for _, gate := range []string{"rear", "source-less", "excluded", "no-shield-flags", "guard", "invalid-stack", "no-update", "dead", "not-npc", "missing-update", "mana-bomb"} {
		t.Run(gate, func(t *testing.T) {
			target, source, attack, shield := npcShieldFixture4E17B0(t)
			ud := target.UpdateDataMonster()
			var events []string
			r := npcShieldRuntime4E17B0(t, target, source, attack, shield, 9, object.DamageImpact, false, &events)
			typ := object.DamageImpact
			switch gate {
			case "rear":
				r.BlockDirection = func(*Object, types.Pointf) bool { return false }
			case "source-less":
				source = nil
			case "excluded":
				r.BlockSourceExcluded = func(*Object) bool { return true }
			case "no-shield-flags":
				ud.ArmorEquipFlags = 0
			case "guard":
				ud.AIStack[1].Action = uint32(ai.ACTION_GUARD)
			case "invalid-stack":
				ud.AIStackInd = -1
			case "no-update":
				target.ObjFlags |= object.FlagNoUpdate
			case "dead":
				target.ObjFlags |= object.FlagDead
			case "not-npc":
				target.ObjSubClass &^= 0x10
			case "missing-update":
				target.UpdateData = nil
			case "mana-bomb":
				typ = object.DamageManaBomb
			}
			before, itemBefore, attackBefore := *ud, *shield, *attack
			var reason string
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			handled, result := PlayerDamageNative4E17B0(target, source, attack, 9, typ, r)
			wantHandled := gate == "no-update" || gate == "dead"
			if handled != wantHandled || result || (!wantHandled && reason == "") || len(events) != 0 || *ud != before || *shield != itemBefore || *attack != attackBefore {
				t.Fatalf("NPC shield gate %s=%t/%t reason=%q effects=%v", gate, handled, result, reason, events)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCShieldDoesNotStopElectric(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		for _, self := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/self-%t", typ, self), func(t *testing.T) {
				target, source, _, shield := npcShieldFixture4E17B0(t)
				// This fixture has no absorbent armor; leave the shield equipped
				// but non-ARMOR so the separate electric durability pass is absent.
				shield.ObjClass = 0
				weapon := (*Object)(nil)
				if self {
					weapon = source
				}
				calls := 0
				r := PlayerDamageRuntime4E17B0{
					BlockDirection:     func(*Object, types.Pointf) bool { t.Fatal("shield tested electric facing"); return true },
					ElectricArmorScale: func(*Object) float32 { return 1 },
					DefaultDamage: func(got, attacker, effective *Object, damage int32, gotType object.DamageType) bool {
						if got != target || attacker != source || effective != weapon || damage != 8 || gotType != typ || target.UpdateDataMonster().Field547 != 2 || target.UpdateDataMonster().Field546 != uint32(typ) {
							t.Fatal("electric hit lost the original tail")
						}
						calls++
						return true
					},
					Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 8, typ, r); !handled || !result || calls != 1 || target.MonsterActionGet50A020() != ai.ACTION_BLOCK_ATTACK {
					t.Fatal("ordinary shield blocked electric damage")
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0NPCShieldSourceOnlyMarkerAndExclusion(t *testing.T) {
	for _, self := range []bool{false, true} {
		for typ := object.DamageType(0); typ <= object.DamageType(18); typ++ {
			if typ == object.DamageElectric || typ == object.DamageAirborneElectric || typ == object.DamageManaBomb {
				continue
			}
			t.Run(fmt.Sprintf("self-%t/type-%d", self, typ), func(t *testing.T) {
				target, source, _, shield := npcShieldFixture4E17B0(t)
				// A cloud-like non-unit source avoids ordinary melee admission.
				source.ObjClass, source.TypeInd = object.ClassSimple, 777
				source.PrevPos, source.PosVec = types.Pointf{X: 20}, types.Pointf{X: -20}
				weapon := (*Object)(nil)
				if self {
					weapon = source
				}
				var events []string
				r := npcShieldRuntime4E17B0(t, target, source, source, shield, 9, typ, false, &events)
				r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
				if !self {
					r.BlockSourceExcluded = func(*Object) bool { t.Fatal("source-only used six weapon exclusions"); return true }
				}
				if self && typ == object.DamageImpact {
					// Restored caster case 11 clears the cached hit marker
					// before facing/audio and selects the live shield after
					// audio/balance, not the old admission-only preflight.
					r.CanDamageBlockItem = func(got *Object) bool {
						if got != shield || target.UpdateDataMonster().Field547 != 0 ||
							!slices.Equal(events, []string{"audio", "percent"}) {
							t.Fatal("caster IMPACT shield admission preceded real block effects")
						}
						return true
					}
				}
				marker, markerType := uint32(0), uint32(77)
				if !self && (typ == object.DamageClaw || typ == object.DamageCrush) {
					marker, markerType = 1, uint32(source.TypeInd)
				}
				r.Audio = func(id int, got *Object) {
					if id != 878 || got != target || target.UpdateDataMonster().Field547 != marker || target.UpdateDataMonster().Field546 != markerType {
						t.Fatal("source-only NPC marker contract")
					}
					events = append(events, "audio")
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, weapon, 9, typ, r); !handled || result || !slices.Equal(events, []string{"audio", "percent", "durability"}) {
					t.Fatalf("source-only NPC block=%t/%t effects=%v", handled, result, events)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0NPCShieldReflectEnchantPrecedesItem(t *testing.T) {
	target, source, attack, shield := npcShieldFixture4E17B0(t)
	target.Buffs = 1 << playerDamageReflectEnchant4E17B0
	attack.PosVec.X = 20
	var events []string
	r := npcReflectRuntime4E17B0(t, target, attack, &events)
	r.DamageBlockItem = func(_, _, _, _ *Object, _ float32, _ object.DamageType) bool {
		t.Fatal("Reflect Shield wore the NPC item")
		return true
	}
	before := *target.UpdateDataMonster()
	if handled, result := PlayerDamageNative4E17B0(target, source, attack, 9, object.DamageImpact, r); !handled || result {
		t.Fatal("Reflect Shield did not precede NPC shield")
	}
	before.Field547 = 0
	if *target.UpdateDataMonster() != before || shield.HealthData.Cur != 20 || !slices.Equal(events, []string{"direction", "reflect", "clear", "set", "audio"}) {
		t.Fatal("Reflect Shield changed NPC shield state")
	}
}

func TestPlayerDamageNative4E17B0NPCShieldMissingItemAndHealth(t *testing.T) {
	for _, shape := range []string{"empty-inventory", "unequipped", "no-match", "no-health"} {
		t.Run(shape, func(t *testing.T) {
			target, source, attack, shield := npcShieldFixture4E17B0(t)
			shield.InvNextItem = nil
			var selected *Object
			switch shape {
			case "empty-inventory":
				target.InvFirstItem = nil
			case "unequipped":
				target.InvFirstItem, shield.ObjFlags = shield, 0
			case "no-match":
				target.InvFirstItem, shield.ObjSubClass = shield, 4
			case "no-health":
				target.InvFirstItem, shield.HealthData, selected = shield, nil, shield
			}
			ud := target.UpdateDataMonster()
			before := *ud
			before.Field547, before.Field546 = 1, uint32(attack.TypeInd)
			var events []string
			r := npcShieldRuntime4E17B0(t, target, source, attack, shield, 9, object.DamageImpact, false, &events)
			attack.ObjSubClass = 0x10 // Non-reflecting flight isolates the wear no-op.
			r.CanDamageBlockItem = func(item *Object) bool {
				if item == nil || item != selected || ud.Field547 != 99 {
					t.Fatal("nil/no-health item preflight identity or prefix")
				}
				return true
			}
			r.DamageBlockItem = func(item, owner, attacker, effective *Object, wear float32, typ object.DamageType) bool {
				if item != selected || owner != target || attacker != source || effective != attack || wear != float32(0.2*9) || typ != object.DamageImpact {
					t.Fatal("nil/no-health item wear arguments")
				}
				events = append(events, "durability")
				// Exercise the real EquipDamage nil/health guard, with no item
				// damage, modifier or health-report service substituted.
				return EquipDamageNative4E16D0(item, owner, attacker, effective, wear, typ, ItemDurabilityDamageRuntime4E1560{})
			}
			r.Melee.MonsterPopBlockAction = func(*Object) { t.Fatal("missing/no-health shield popped the NPC action") }
			if selected == nil {
				r.Melee.MonsterPopBlockAction = nil
			}
			if handled, result := PlayerDamageNative4E17B0(target, source, attack, 9, object.DamageImpact, r); !handled || result ||
				!slices.Equal(events, []string{"audio", "percent", "durability"}) || *ud != before || target.HealthData.Cur != 60 {
				t.Fatalf("missing/no-health shield=%t/%t events=%v", handled, result, events)
			}
		})
	}
}
