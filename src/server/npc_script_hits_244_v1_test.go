package server

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// The matrix is also checked against unchanged GAME.EXE 004E17B0 -> 004E0B30.
// Case 2 uses half absorption; case 8 uses full absorption. Neither nil-source
// case may dereference a source, acquire a weapon marker, or run facing/shield
// stance logic. Existing ordinary-monster admission tests are left unchanged.
func TestNPCScriptHits244V1Matrix(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageBite} {
		for _, raw := range []int32{-9, -1, 0, 1, 3, 9, 50, 200} {
			for _, absorption := range []float32{0, .25, .5, 1, 1.5} {
				for _, carry := range []float32{-.5, .25, .5} {
					for _, mode := range []struct {
						quest bool
						scale float32
					}{{false, 1}, {true, -1}, {true, 0}, {true, .5}, {true, 1.5}} {
						t.Run(fmt.Sprintf("type-%d/raw-%d/armor-%g/carry-%g/quest-%t-%g", typ, raw, absorption, carry, mode.quest, mode.scale), func(t *testing.T) {
							npc := damageMeleeUnitFixture4E17B0(t, false)
							npc.TypeInd, npc.ObjClass, npc.ObjSubClass, npc.ObjFlags = 1343, 2097154, 69650, 16777732
							*npc.HealthData = HealthData{Cur: 1000, Field2: 1000, Max: 1000}
							ud := npc.UpdateDataMonster()
							ud.Field518, ud.Field1 = math.Float32bits(absorption), math.Float32bits(carry)
							ud.Field547, ud.Field546 = 99, 77
							npc.Obj130, npc.Field131, npc.Frame134 = &Object{}, 999, 777
							npc.Pos132 = types.Ptf(123, -123)
							npc.Buffs = 1 << playerDamageReflectEnchant4E17B0
							original := sourceLessPhysicalHistoricalCase4E17B0(t, "NPC244-matrix", false, typ, raw, absorption, carry, mode.quest, mode.scale)
							want, wantCarry := int32(original.ExpectedDamage[0]), original.ExpectedCarry
							world := damageMeleeWorldRuntime4E0B30(t)
							health, defaults, gods, scales := 0, 0, 0, 0
							world.BuffOff = func(*Object, EnchantID) { t.Fatal("nil hit cleared invisibility") }
							world.DamageClear = func(target *Object, amount int32) {
								if target != npc || amount != want {
									t.Fatalf("health amount=%d want=%d", amount, want)
								}
								health++
								target.HealthData.Cur = uint16(1000 - amount)
							}
							r := PlayerDamageRuntime4E17B0{
								BlockSourceExcluded: func(*Object) bool { t.Fatal("nil hit exclusion"); return false },
								BlockDirection:      func(*Object, types.Pointf) bool { t.Fatal("nil hit facing"); return false },
								ProjectileReflect:   func(*Object, *Object) { t.Fatal("nil hit reflection") },
								GodMode: func() bool {
									gods++
									if ud.Field547 != 2 || ud.Field546 != uint32(typ) {
										t.Fatal("GodMode before cached marker")
									}
									return true // NPC is not made immune by player-only GodMode.
								},
								QuestMode:        func() bool { return mode.quest },
								QuestDamageScale: func() float32 { scales++; return mode.scale },
								DefaultDamage: func(target, source, weapon *Object, amount int32, damageType object.DamageType) bool {
									if target != npc || source != nil || weapon != nil || amount != want || damageType != typ || ud.Field1 != wantCarry {
										t.Fatalf("default amount=%d want=%d carry=%x/%x", amount, want, ud.Field1, wantCarry)
									}
									defaults++
									return DefaultDamageWorld4E0B30(target, source, weapon, amount, damageType, world)
								},
								Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
							}
							h, result := PlayerDamageNative4E17B0(npc, nil, nil, raw, typ, r)
							if !h || !result || health != 1 || defaults != 1 || gods != 1 || scales != map[bool]int{false: 0, true: 1}[mode.quest] ||
								ud.Field1 != wantCarry || ud.Field547 != 2 || ud.Field546 != uint32(typ) || npc.Obj130 != nil || npc.Pos132 != (types.Pointf{}) ||
								npc.Field131 != uint32(typ) || npc.Frame134 != 1400 || !ud.StatusFlags.Has(object.MonStatusInjured) || npc.HealthData.Cur != uint16(1000-want) {
								t.Fatalf("handled/result=%t/%t calls=%d/%d/%d/%d HP=%d carry=%x marker=%d/%d", h, result, health, defaults, gods, scales, npc.HealthData.Cur, ud.Field1, ud.Field547, ud.Field546)
							}
							row, err := json.Marshal(map[string]any{"typ": typ, "damage": raw, "absorption": absorption, "carry": carry, "quest": mode.quest, "scale": mode.scale,
								"result": 1, "carry_bits": ud.Field1, "marker": ud.Field547, "marker_type": ud.Field546, "status": ud.StatusFlags,
								"attacker": 0, "damage_type": npc.Field131, "frame": npc.Frame134, "position": [2]int{0, 0}, "hp": npc.HealthData.Cur})
							if err != nil {
								t.Fatal(err)
							}
							t.Logf("R244 NPC HIT: %s", row)
						})
					}
				}
			}
		}
	}
}

func TestNPCScriptHits244V1ArmorLiveCarryAndMarker(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageBite} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			npc := damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(npc, .5, .4)
			*(*float32)(armor.UpdateData) = 0
			cached := npc.UpdateDataMonster()
			cached.Field547, cached.Field546 = 99, 77
			live := &MonsterUpdateData{Field518: math.Float32bits(.25), Field1: math.Float32bits(.25), Field547: 33, Field546: 44}
			wantEffective, wantCarry, wantWear, wantDefault := int32(4), float32(.75), int32(5), int32(2)
			if typ == object.DamageCrush {
				wantEffective, wantCarry, wantWear, wantDefault = 7, 0, 4, 3 // raw9-effective7, live armor .25 / live denominator .25.
			}
			var events []string
			r := PlayerDamageRuntime4E17B0{
				CanDamageArmor: func(item *Object) bool { npc.UpdateData = unsafe.Pointer(live); return item == armor },
				ItemArmorValue: func(item *Object) float32 {
					if item != armor || cached.Field547 != 0 || live.Field1 != math.Float32bits(wantCarry) {
						t.Fatal("armor ran before live carry / marker reset")
					}
					events = append(events, "lookup")
					return .5 // live denominator .25 produces factor two.
				},
				DamageArmor: func(item, source, weapon *Object, amount int32, got object.DamageType) bool {
					if typ == object.DamageBite {
						wantWear = 10
					}
					if item != armor || source != nil || weapon != nil || amount != wantWear || got != typ {
						t.Fatalf("armor amount=%d want=%d effective=%d", amount, wantWear, wantEffective)
					}
					events = append(events, "wear")
					cached.Field547, cached.Field546 = 1, 888
					return true
				},
				GodMode:          func() bool { events = append(events, "god"); return true },
				QuestMode:        func() bool { events = append(events, "quest"); return true },
				QuestDamageScale: func() float32 { events = append(events, "scale"); return .5 },
				DefaultDamage: func(target, source, weapon *Object, amount int32, got object.DamageType) bool {
					if target != npc || source != nil || weapon != nil || amount != wantDefault || got != typ || cached.Field547 != 1 || cached.Field546 != 888 || live.Field547 != 33 || live.Field546 != 44 {
						t.Fatal("cached marker/live carry or late Quest order")
					}
					events = append(events, "default")
					return true
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			if h, result := PlayerDamageNative4E17B0(npc, nil, nil, 9, typ, r); !h || !result || cached.Field1 != math.Float32bits(.4) || live.Field1 != math.Float32bits(wantCarry) ||
				!slices.Equal(events, []string{"lookup", "wear", "god", "quest", "scale", "default"}) {
				t.Fatalf("live armor/carry events=%v cached=%x live=%x", events, cached.Field1, live.Field1)
			}
		})
	}
}

func TestNPCScriptHits244V1GatesAndLateDefense(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageBite} {
		for _, gate := range []string{"defense-stance-with-nil-source", "no-update", "dead", "invulnerable", "late-defend-shield", "shield-absorbs-all"} {
			t.Run(fmt.Sprintf("type-%d/%s", typ, gate), func(t *testing.T) {
				npc := damageMeleeUnitFixture4E17B0(t, false)
				ud := npc.UpdateDataMonster()
				ud.Field518, ud.Field1 = math.Float32bits(.5), math.Float32bits(.25)
				ud.Field547, ud.Field546 = 99, 77
				ud.ArmorEquipFlags, ud.WeaponEquipFlags = 0x3000000, 0x400
				ud.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
				npc.Buffs = 1 << playerDamageReflectEnchant4E17B0
				switch gate {
				case "no-update":
					npc.ObjFlags |= object.FlagNoUpdate
				case "dead":
					npc.ObjFlags |= object.FlagDead
				case "invulnerable":
					npc.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "late-defend-shield", "shield-absorbs-all":
					npc.Buffs |= 1 << playerDamageShieldEnchant4E17B0
				}
				before, buffs := *ud, npc.Buffs
				world := damageMeleeWorldRuntime4E0B30(t)
				var events []string
				want := int32(25)
				if typ == object.DamageCrush {
					want = 37
				}
				world.BuffOff = func(*Object, EnchantID) { t.Fatal("nil hit removed buff") }
				if gate == "late-defend-shield" {
					modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
					init := &ModifierInitData{}
					init.Modifiers[2] = modifier
					npc.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
					world.CanApplyLateDefend = func(got *ModifierEff) bool { return got == modifier }
					world.ApplyLateDefend = func(got *ModifierEff, _, target, weapon, source *Object, amount int32, gotType object.DamageType) int32 {
						if got != modifier || target != npc || source != nil || weapon != nil || amount != want || gotType != typ {
							t.Fatal("nil late defense arguments")
						}
						events = append(events, "defend")
						return amount - 3
					}
				}
				world.DefaultDamageSound = func(target, source *Object) {
					if target != npc || source != nil || npc.Field131 != uint32(typ) || npc.Frame134 != 1400 || ud.Field547 != 2 {
						t.Fatal("sound attribution/order")
					}
					events = append(events, "sound")
				}
				world.ShieldReduce = func(target *Object, amount *int32, got object.DamageType, source *Object) {
					if target != npc || got != typ || source != nil {
						t.Fatal("nil Shield arguments")
					}
					events = append(events, "shield")
					if gate == "shield-absorbs-all" {
						*amount = 0
					} else {
						*amount -= 2
					}
				}
				health := world.DamageClear
				world.DamageClear = func(target *Object, amount int32) { events = append(events, "health"); health(target, amount) }
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("nil hit facing/stance"); return true }
				r.ProjectileReflect = func(*Object, *Object) { t.Fatal("nil Reflect") }
				r.Audio = func(id int, target *Object) {
					if gate != "invulnerable" || id != 71 || target != npc {
						t.Fatal("unexpected audio")
					}
					events = append(events, "audio")
				}
				r.DefaultDamage = func(target, source, weapon *Object, amount int32, got object.DamageType) bool {
					return DefaultDamageWorld4E0B30(target, source, weapon, amount, got, world)
				}
				h, result := PlayerDamageNative4E17B0(npc, nil, nil, 50, typ, r)
				if !h || npc.Buffs != buffs {
					t.Fatal("gate/buff mutation")
				}
				switch gate {
				case "no-update", "dead":
					if result || npc.HealthData.Cur != 200 || *ud != before || len(events) != 0 {
						t.Fatal("NoUpdate/Dead gate moved")
					}
				case "invulnerable":
					if !result || npc.HealthData.Cur != 200 || *ud != before || !slices.Equal(events, []string{"audio"}) {
						t.Fatal("invulnerable gate moved")
					}
				case "shield-absorbs-all":
					if result || npc.HealthData.Cur != 200 || !slices.Equal(events, []string{"sound", "shield"}) {
						t.Fatal("Shield zero return/HP")
					}
				case "late-defend-shield":
					if !result || npc.HealthData.Cur != uint16(200-want+5) || !slices.Equal(events, []string{"defend", "sound", "shield", "health"}) {
						t.Fatalf("late effects HP=%d events=%v", npc.HealthData.Cur, events)
					}
				default:
					if !result || npc.HealthData.Cur != uint16(200-want) || ud.Field547 != 2 || ud.Field546 != uint32(typ) || !slices.Equal(events, []string{"sound", "health"}) {
						t.Fatal("nil source was blocked by stance")
					}
				}
			})
		}
	}
}

func TestNPCScriptHits244V1AdmissionBoundary(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageBite} {
		for _, boundary := range []string{"ordinary-monster", "missing-update", "nil-source-weapon", "world-source-no-weapon"} {
			t.Run(fmt.Sprintf("type-%d/%s", typ, boundary), func(t *testing.T) {
				npc := damageMeleeUnitFixture4E17B0(t, false)
				var source, weapon *Object
				switch boundary {
				case "ordinary-monster":
					npc.ObjSubClass = 0x202
				case "missing-update":
					npc.UpdateData = nil
				case "nil-source-weapon":
					weapon = &Object{}
				case "world-source-no-weapon":
					source = &Object{}
				}
				rejects := 0
				h, result := PlayerDamageNative4E17B0(npc, source, weapon, 200, typ, PlayerDamageRuntime4E17B0{
					DefaultDamage: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
						t.Fatal("unrelated damage admitted")
						return true
					},
					Unsupported: func(_ string, _, _, _ *Object, _ int32, _ object.DamageType) { rejects++ },
				})
				if h || result || rejects != 1 || npc.HealthData.Cur != 200 {
					t.Fatal("unrelated shape boundary changed")
				}
			})
		}
	}
}
