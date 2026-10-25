package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestPlayerDamageNPCScriptBlade4E17B0Wiz01aLethal(t *testing.T) {
	npc := damageMeleeUnitFixture4E17B0(t, false)
	ud := npc.UpdateDataMonster()
	ud.Field547, ud.Field546 = 99, 77
	r := damageMeleeRuntimeFixture4E17B0(t)
	h, result := PlayerDamageNative4E17B0(npc, nil, nil, 500, object.DamageBlade, r)
	if !h || !result || npc.HealthData.Cur != 0 || ud.Field547 != 2 || ud.Field546 != 0 {
		t.Fatalf("Wiz01a Damage(Apprentice2, nil, 500, BLADE): handled=%t result=%t HP=%d marker=%d/%d", h, result, npc.HealthData.Cur, ud.Field547, ud.Field546)
	}
}

func TestPlayerDamageNPCScriptBlade4E17B0SignedArmorQuest(t *testing.T) {
	for _, raw := range []int32{-9, -1, 0, 1, 3, 9, 500} {
		for _, absorption := range []float32{0, .25, .5, 1} {
			for _, initialCarry := range []float32{-.5, .25, .5} {
				for _, scale := range []float32{-1, 0, .5, 1.5} {
					t.Run(fmt.Sprintf("raw-%d/armor-%g/carry-%g/quest-%g", raw, absorption, initialCarry, scale), func(t *testing.T) {
						npc := damageMeleeUnitFixture4E17B0(t, false)
						armor := damageMeleeArmorFixture4E17B0(npc, absorption, initialCarry)
						*(*float32)(armor.UpdateData) = 0
						if absorption == 0 {
							npc.InvFirstItem = nil
						}
						ud := npc.UpdateDataMonster()
						ud.Field547, ud.Field546 = 99, 77
						// Expected numbers come from actual original startup-control execution.
						original := sourceLessPhysicalHistoricalCase4E17B0(t, "NPCBlade-matrix", false, object.DamageBlade, raw, absorption, initialCarry, scale >= 0, scale)
						carry, wear, want := math.Float32frombits(original.ExpectedCarry), int32(original.ExpectedWear), int32(original.ExpectedDamage[0])
						lookups, wears, defaults, gods := 0, 0, 0, 0
						r := PlayerDamageRuntime4E17B0{
							BlockSourceExcluded: func(*Object) bool { t.Fatal("nil source ran exclusion"); return false },
							BlockDirection:      func(*Object, types.Pointf) bool { t.Fatal("nil source ran facing"); return false },
							ProjectileReflect:   func(*Object, *Object) { t.Fatal("nil source ran Reflect") },
							ItemArmorValue: func(item *Object) float32 {
								if item != armor || ud.Field547 != 0 {
									t.Fatal("armor lookup before marker reset")
								}
								lookups++
								return absorption
							},
							CanDamageArmor: func(item *Object) bool { return item == armor },
							DamageArmor: func(item, source, weapon *Object, amount int32, typ object.DamageType) bool {
								if item != armor || source != nil || weapon != nil || amount != wear || typ != object.DamageBlade || ud.Field1 != math.Float32bits(carry) {
									t.Fatalf("armor amount=%d/%d carry=%x/%x attribution=%p/%p", amount, wear, ud.Field1, math.Float32bits(carry), source, weapon)
								}
								wears++
								return true
							},
							GodMode: func() bool {
								if ud.Field547 != 2 || ud.Field546 != 0 {
									t.Fatal("GodMode before cached marker")
								}
								gods++
								return true
							},
							QuestMode:        func() bool { return scale >= 0 },
							QuestDamageScale: func() float32 { return scale },
							DefaultDamage: func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
								if target != npc || source != nil || weapon != nil || amount != want || typ != object.DamageBlade {
									t.Fatalf("default damage=%d/%d target=%p/%p attribution=%p/%p", amount, want, target, npc, source, weapon)
								}
								defaults++
								return true
							},
							Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
						}
						h, result := PlayerDamageNative4E17B0(npc, nil, nil, raw, object.DamageBlade, r)
						wantLookups, wantWear := 0, 0
						if absorption > 0 {
							wantLookups = 1
							if wear > 0 {
								wantWear = 1
							}
						}
						if !h || !result || defaults != 1 || gods != 1 || lookups != wantLookups || wears != wantWear || ud.Field1 != math.Float32bits(carry) {
							t.Fatalf("handled=%t result=%t default=%d god=%d lookup=%d/%d wear=%d/%d carry=%x/%x", h, result, defaults, gods, lookups, wantLookups, wears, wantWear, ud.Field1, math.Float32bits(carry))
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNPCScriptBlade4E17B0DefenseGates(t *testing.T) {
	for _, mode := range []string{"nil-source-defense", "invulnerable", "dead", "no-update"} {
		t.Run(mode, func(t *testing.T) {
			npc := damageMeleeUnitFixture4E17B0(t, false)
			ud := npc.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			ud.ArmorEquipFlags, ud.WeaponEquipFlags = 0x3000000, 0x400
			ud.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			npc.Buffs = 1 << playerDamageReflectEnchant4E17B0
			switch mode {
			case "invulnerable":
				npc.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			case "dead":
				npc.ObjFlags |= object.FlagDead
			case "no-update":
				npc.ObjFlags |= object.FlagNoUpdate
			}
			before, buffs := *ud, npc.Buffs
			r := damageMeleeRuntimeFixture4E17B0(t)
			sounds := 0
			r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("source-less defense"); return true }
			r.ProjectileReflect = func(*Object, *Object) { t.Fatal("source-less Reflect") }
			r.Audio = func(id int, target *Object) {
				if mode != "invulnerable" || id != playerDamageInvulnerableSound4E17B0 || target != npc {
					t.Fatal("unexpected defense audio")
				}
				sounds++
			}
			h, result := PlayerDamageNative4E17B0(npc, nil, nil, 9, object.DamageBlade, r)
			if !h || npc.Buffs != buffs {
				t.Fatal("source-less NPC gate/buff mutation")
			}
			if mode == "nil-source-defense" {
				if !result || npc.HealthData.Cur != 191 || ud.Field547 != 2 || ud.Field546 != 0 || sounds != 0 {
					t.Fatal("source-less blade was blocked by a defense stance")
				}
			} else if npc.HealthData.Cur != 200 || *ud != before || (mode == "invulnerable") != result || sounds != map[bool]int{true: 1, false: 0}[mode == "invulnerable"] {
				t.Fatalf("gate=%s handled=%t result=%t HP=%d sounds=%d", mode, h, result, npc.HealthData.Cur, sounds)
			}
		})
	}
}

func TestPlayerDamageNPCScriptBlade4E17B0CachedAbsorptionLiveCarryMarker(t *testing.T) {
	npc := damageMeleeUnitFixture4E17B0(t, false)
	armor := damageMeleeArmorFixture4E17B0(npc, .5, .4)
	*(*float32)(armor.UpdateData) = 0
	cached := npc.UpdateDataMonster()
	cached.Field547, cached.Field546 = 99, 77
	live := &MonsterUpdateData{Field518: math.Float32bits(.25), Field1: math.Float32bits(.25), Field547: 33, Field546: 44}
	var events []string
	r := PlayerDamageRuntime4E17B0{
		CanDamageArmor: func(item *Object) bool {
			if item != armor {
				t.Fatal("armor admission")
			}
			npc.UpdateData = unsafe.Pointer(live)
			return true
		},
		ItemArmorValue: func(item *Object) float32 {
			if item != armor || cached.Field547 != 0 || live.Field1 != math.Float32bits(.75) {
				t.Fatal("cached absorption/live carry/marker reset")
			}
			events = append(events, "lookup")
			return .25
		},
		DamageArmor: func(item, source, weapon *Object, amount int32, typ object.DamageType) bool {
			if item != armor || source != nil || weapon != nil || amount != 5 || typ != object.DamageBlade {
				t.Fatal("live armor denominator")
			}
			events = append(events, "wear")
			cached.Field547, cached.Field546 = 1, 888
			return true
		},
		GodMode:          func() bool { events = append(events, "god"); return true },
		QuestMode:        func() bool { events = append(events, "quest"); return true },
		QuestDamageScale: func() float32 { events = append(events, "scale"); return .5 },
		DefaultDamage: func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
			if target != npc || source != nil || weapon != nil || amount != 2 || typ != object.DamageBlade || cached.Field547 != 1 || cached.Field546 != 888 || live.Field547 != 33 || live.Field546 != 44 {
				t.Fatal("cached marker callback overwrite or late Quest scaling")
			}
			events = append(events, "default")
			return true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
	}
	if h, result := PlayerDamageNative4E17B0(npc, nil, nil, 9, object.DamageBlade, r); !h || !result {
		t.Fatal("live update replacement rejected")
	}
	if cached.Field1 != math.Float32bits(.4) || live.Field1 != math.Float32bits(.75) || !slices.Equal(events, []string{"lookup", "wear", "god", "quest", "scale", "default"}) {
		t.Fatalf("carry/order: cached=%x live=%x events=%v", cached.Field1, live.Field1, events)
	}
}

func TestPlayerDamageNPCScriptBlade4E17B0RejectsUnrelatedShapes(t *testing.T) {
	for _, mode := range []string{"player", "ordinary-monster", "missing-update", "distinct-nil-source-weapon", "source-no-weapon"} {
		t.Run(mode, func(t *testing.T) {
			npc := damageMeleeUnitFixture4E17B0(t, mode == "player")
			var source, weapon *Object
			switch mode {
			case "ordinary-monster":
				npc.ObjSubClass = 0x202
			case "missing-update":
				npc.UpdateData = nil
			case "distinct-nil-source-weapon":
				weapon = &Object{}
			case "source-no-weapon":
				source = &Object{}
			}
			rejected, defaults := 0, 0
			r := PlayerDamageRuntime4E17B0{
				DefaultDamage: func(target, a, w *Object, d int32, typ object.DamageType) bool {
					if target != npc || a != nil || w != nil || d != 500 || typ != object.DamageBlade {
						t.Fatal("unexpected script BLADE default boundary")
					}
					defaults++
					return true
				},
				Unsupported: func(_ string, _, _, _ *Object, _ int32, _ object.DamageType) { rejected++ },
			}
			h, result := PlayerDamageNative4E17B0(npc, source, weapon, 500, object.DamageBlade, r)
			wantHandled, wantRejected, wantDefaults := false, 1, 0
			if mode == "player" {
				// Player source-less BLADE is now ported too. Keep this exact
				// input as a positive control; the other four stay rejected.
				wantHandled, wantRejected, wantDefaults = true, 0, 1
			}
			if h != wantHandled || result != wantHandled || rejected != wantRejected || defaults != wantDefaults || npc.HealthData.Cur != 200 {
				t.Fatalf("unrelated=%s handled=%t result=%t reject=%d default=%d HP=%d", mode, h, result, rejected, defaults, npc.HealthData.Cur)
			}
		})
	}
}
