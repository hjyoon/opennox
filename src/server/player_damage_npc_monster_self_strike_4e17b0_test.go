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

var npcMonsterSelfStrikeTypes4E17B0 = []object.DamageType{
	object.DamageBlade, object.DamageCrush, object.DamageImpale,
	object.DamageDrain, object.DamageBite, object.DamageClaw,
}

func TestPlayerDamageNPCMonsterSelfStrikeArmor4E17B0(t *testing.T) {
	for _, typ := range npcMonsterSelfStrikeTypes4E17B0 {
		for _, absorption := range []float32{0, 0.25, 1} {
			for _, raw := range []int32{-3, 0, 5, 40} {
				for _, mode := range []string{"normal", "quest", "god"} {
					t.Run(fmt.Sprintf("type-%d/armor-%g/raw-%d/%s", typ, absorption, raw, mode), func(t *testing.T) {
						v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
						ud := v.UpdateDataMonster()
						ud.Field518, ud.Field1 = math.Float32bits(absorption), math.Float32bits(0.125)
						ud.Field547, ud.Field546 = 99, 77
						r := damageMeleeRuntimeFixture4E17B0(t)
						var events []string
						r.BlockSourceExcluded = func(attack *Object) bool {
							if attack != a || ud.Field547 != 0 || ud.Field546 != 77 {
								t.Fatal("NPC self-strike marker must clear before exclusions")
							}
							events = append(events, "exclude")
							return false
						}
						r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "facing"); return false }
						x := float64(absorption)
						if typ == object.DamageCrush {
							x *= 0.5
						}
						accumulated := float32((1-x)*float64(raw)) + 0.125
						rounded := int32(math.RoundToEven(float64(accumulated)))
						effective, carry := rounded, accumulated-float32(rounded)
						if typ == object.DamageDrain {
							effective, carry = raw, 0.125
						}
						if raw > 0 && effective == 0 {
							effective = 1
						}
						if mode == "quest" {
							before := effective
							effective = int32(math.RoundToEven(float64(float32(float64(effective) * 0.5))))
							if before > 0 && effective < 1 {
								effective = 1
							}
						}
						r.GodMode = func() bool {
							if ud.Field547 != 2 || ud.Field546 != uint32(typ) || ud.Field1 != math.Float32bits(carry) {
								t.Fatal("NPC carry/marker precede the player-only GodMode gate")
							}
							events = append(events, "god")
							return mode == "god"
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return mode == "quest" }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
						r.DefaultDamage = func(target, source, weapon *Object, amount int32, kind object.DamageType) bool {
							if target != v || source != a || weapon != a || amount != effective || kind != typ {
								t.Fatalf("NPC default type=%d amount=%d want=%d", kind, amount, effective)
							}
							events = append(events, "default")
							return true
						}
						if h, result := PlayerDamageNative4E17B0(v, a, a, raw, typ, r); !h || !result {
							t.Fatal("NPC self-strike rejected")
						}
						want := []string{"exclude", "facing", "god", "quest"}
						if mode == "quest" {
							want = append(want, "scale")
						}
						want = append(want, "default")
						if !slices.Equal(events, want) || v.HealthData.Cur != 200 {
							t.Fatalf("events=%v want=%v", events, want)
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNPCMonsterSelfStrikeCachedPrefix4E17B0(t *testing.T) {
	for _, typ := range npcMonsterSelfStrikeTypes4E17B0 {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/replace-%t", typ, replace), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
				cached, live := v.UpdateDataMonster(), v.UpdateDataMonster()
				cached.Field547, cached.Field546 = 99, 77
				armor := damageMeleeArmorFixture4E17B0(v, 0.25, 0.125)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.875)
				pos := a.PrevPos
				exclusions, directions, wear, armorLookups := 0, 0, 0, 0
				r.ItemArmorValue = func(*Object) float32 { armorLookups++; return 0.875 }
				x := 0.25
				if typ == object.DamageCrush {
					x *= 0.5
				}
				accumulated := float32((1-x)*5) + 0.5
				effective := int32(math.RoundToEven(float64(accumulated)))
				carry := accumulated - float32(effective)
				if typ == object.DamageDrain {
					effective, carry = 5, 0.5
				}
				r.BlockSourceExcluded = func(*Object) bool {
					if cached.Field547 != 0 || cached.Field546 != 77 {
						t.Fatal("self-weapon must not be attributed as a separate projectile")
					}
					if replace {
						live = &MonsterUpdateData{Field547: 55}
						v.UpdateData = unsafe.Pointer(live)
					}
					live.Field518, live.Field1 = math.Float32bits(0.875), math.Float32bits(0.5)
					a.PrevPos = types.Ptf(-100, 99)
					exclusions++
					return false
				}
				r.BlockDirection = func(_ *Object, p types.Pointf) bool {
					if p != pos || cached.Field547 != 0 {
						t.Fatal("NPC PrevPos must be snapshotted before callbacks")
					}
					directions++
					return false
				}
				r.DamageArmor = func(_, source, weapon *Object, amount int32, kind object.DamageType) bool {
					if source != a || weapon != a || amount != 5-effective || kind != typ || cached.Field547 != 0 || live.Field1 != math.Float32bits(carry) || typ == object.DamageDrain {
						t.Fatal("NPC live carry/armor wear ordering")
					}
					wear++
					return true
				}
				r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
					if source != a || weapon != a || amount != effective || cached.Field547 != 2 || cached.Field546 != uint32(typ) {
						t.Fatal("NPC cached absorption/marker tail")
					}
					return true
				}
				wantWear, wantLookups := 1, 1
				if typ == object.DamageDrain || effective == 5 {
					wantWear = 0
				}
				if typ == object.DamageDrain {
					wantLookups = 0
				}
				if h, result := PlayerDamageNative4E17B0(v, a, a, 5, typ, r); !h || !result || exclusions != 1 || directions != 1 || wear != wantWear || armorLookups != wantLookups || live.Field1 != math.Float32bits(carry) || (replace && (live.Field547 != 55 || cached.Field1 != math.Float32bits(0.125))) {
					t.Fatalf("NPC cached/live contract: exclusions=%d directions=%d wear=%d", exclusions, directions, wear)
				}
			})
		}
	}
}

func TestPlayerDamageNPCMonsterSelfStrikeDefense4E17B0(t *testing.T) {
	for _, typ := range npcMonsterSelfStrikeTypes4E17B0 {
		for _, equipment := range []string{"shield", "sword", "staff"} {
			for _, front := range []bool{false, true} {
				for _, excluded := range []bool{false, true} {
					t.Run(fmt.Sprintf("type-%d/%s/front-%t/excluded-%t", typ, equipment, front, excluded), func(t *testing.T) {
						v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
						ud := v.UpdateDataMonster()
						ud.Field547, ud.Field546 = 99, 77
						mask, sound := uint32(2), 878
						if equipment == "shield" {
							ud.ArmorEquipFlags = 0x1000000
							ud.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
						} else if equipment == "sword" {
							ud.WeaponEquipFlags, mask, sound = 0x400, 0x400, 890
						} else {
							ud.WeaponEquipFlags, mask, sound = 0x8000, 0x8000, 894
						}
						item := &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 20, Max: 20}}
						v.InvFirstItem = item
						r := damageMeleeRuntimeFixture4E17B0(t)
						r.BlockSourceExcluded = func(*Object) bool { return excluded }
						r.BlockDirection = func(*Object, types.Pointf) bool {
							if excluded {
								t.Fatal("excluded strike reached facing test")
							}
							return front
						}
						var events []string
						r.Audio = func(id int, _ *Object) {
							if id != sound || ud.Field547 != 0 {
								t.Fatal("NPC block audio/marker")
							}
							events = append(events, "audio")
						}
						r.Melee.MonsterBlockAction = func(*Object) { events = append(events, "action") }
						r.BlockDamagePercent = func() float64 { events = append(events, "percent"); return 0.25 }
						r.CanDamageBlockItem = func(*Object) bool { return true }
						r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
						wear := func(selected, owner, attacker, attack *Object, amount float32, kind object.DamageType) bool {
							if selected != item || owner != v || attacker != a || attack != a || amount != 10 || kind != typ || ud.Field547 != 0 {
								t.Fatal("NPC block durability identity")
							}
							events = append(events, "wear")
							return true
						}
						r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
						blocked := front && !excluded && (equipment == "shield" || typ == object.DamageBlade)
						r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
							if blocked || source != a || weapon != a || amount != 40 || ud.Field547 != 2 || ud.Field546 != uint32(typ) {
								t.Fatal("NPC unblocked tail")
							}
							events = append(events, "default")
							return true
						}
						if h, result := PlayerDamageNative4E17B0(v, a, a, 40, typ, r); !h || result == blocked {
							t.Fatalf("NPC defense=%t/%t blocked=%t", h, result, blocked)
						}
						want := []string{"default"}
						if blocked {
							want = []string{"audio"}
							if equipment != "shield" {
								want = append(want, "action")
							}
							want = append(want, "percent", "wear")
						}
						if !slices.Equal(events, want) {
							t.Fatalf("events=%v want=%v", events, want)
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamageNPCMonsterSelfStrikeActualHP4E17B0(t *testing.T) {
	for _, typ := range npcMonsterSelfStrikeTypes4E17B0 {
		for _, absorption := range []float32{0, 0.25} {
			t.Run(fmt.Sprintf("type-%d/armor-%g", typ, absorption), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
				v.UpdateDataMonster().Field518 = math.Float32bits(absorption)
				r := damageMeleeRuntimeFixture4E17B0(t)
				x := float64(absorption)
				if typ == object.DamageCrush {
					x *= 0.5
				}
				effective := int32(math.RoundToEven(float64(float32((1 - x) * 10))))
				if typ == object.DamageDrain {
					effective = 10
				}
				if h, result := PlayerDamageNative4E17B0(v, a, a, 10, typ, r); !h || !result {
					t.Fatal("NPC actual HP path rejected")
				}
				ud := v.UpdateDataMonster()
				if v.HealthData.Cur != uint16(200-effective) || ud.Field547 != 2 || ud.Field546 != uint32(typ) || !ud.StatusFlags.Has(object.MonStatusInjured) || v.Obj130 != a || v.Frame134 != 1400 || a.UpdateDataMonster().Field130 != 1400 {
					t.Fatalf("NPC HP/Injured/attribution: type=%d HP=%d want=%d", typ, v.HealthData.Cur, 200-effective)
				}
			})
		}
	}
}
