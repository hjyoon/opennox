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

func TestPlayerDamageNPCMonsterSelfImpactArmor4E17B0(t *testing.T) {
	for _, absorption := range []float32{0, 0.25, 1} {
		for _, raw := range []int32{-3, 0, 5, 40} {
			for _, mode := range []string{"normal", "quest", "god"} {
				t.Run(fmt.Sprintf("armor-%g/raw-%d/%s", absorption, raw, mode), func(t *testing.T) {
					v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
					ud := v.UpdateDataMonster()
					ud.Field518, ud.Field1 = math.Float32bits(absorption), math.Float32bits(0.125)
					ud.Field547, ud.Field546 = 99, 77
					r := damageMeleeRuntimeFixture4E17B0(t)
					var events []string
					r.BlockSourceExcluded = func(attack *Object) bool {
						if attack != a || ud.Field547 != 0 || ud.Field546 != 77 {
							t.Fatal("NPC prefix not cleared before exclusions")
						}
						events = append(events, "exclude")
						return false
					}
					r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "facing"); return false }
					accumulated := float32((1-float64(absorption))*float64(raw)) + 0.125
					rounded := playerDamageRound4E17B0(accumulated)
					effective := rounded
					if raw > 0 && effective == 0 {
						effective = 1
					}
					if mode == "quest" {
						before := effective
						effective = playerDamageRound4E17B0(float32(float64(effective) * 0.5))
						if before > 0 && effective < 1 {
							effective = 1
						}
					}
					r.GodMode = func() bool {
						if ud.Field547 != 2 || ud.Field546 != 11 || ud.Field1 != math.Float32bits(accumulated-float32(rounded)) {
							t.Fatal("NPC God read preceded carry/type")
						}
						events = append(events, "god")
						return mode == "god"
					}
					r.QuestMode = func() bool { events = append(events, "quest"); return mode == "quest" }
					r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
					r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != a || amount != effective || typ != object.DamageImpact {
							t.Fatalf("NPC default amount=%d want=%d", amount, effective)
						}
						events = append(events, "default")
						return true
					}
					if h, result := PlayerDamageNative4E17B0(v, a, a, raw, object.DamageImpact, r); !h || !result {
						t.Fatal("NPC self-weapon IMPACT rejected")
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

func TestPlayerDamageNPCMonsterSelfImpactCachedPrefix4E17B0(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(fmt.Sprintf("replace-%t", replace), func(t *testing.T) {
			v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
			cached, live := v.UpdateDataMonster(), v.UpdateDataMonster()
			cached.Field547, cached.Field546 = 99, 77
			armor := damageMeleeArmorFixture4E17B0(v, 0.25, 0.125)
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.875)
			pos := a.PrevPos
			exclusions, directions, wear := 0, 0, 0
			r.BlockSourceExcluded = func(*Object) bool {
				if cached.Field547 != 0 || cached.Field546 != 77 {
					t.Fatal("NPC self weapon was attributed as distinct")
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
					t.Fatal("NPC PrevPos/marker snapshot")
				}
				directions++
				return false
			}
			r.DamageArmor = func(_, source, weapon *Object, amount int32, typ object.DamageType) bool {
				if source != a || weapon != a || amount != 1 || typ != object.DamageImpact || cached.Field547 != 0 || live.Field1 != math.Float32bits(0.25) {
					t.Fatal("NPC carry/armor ordering")
				}
				wear++
				return true
			}
			r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
				if source != a || weapon != a || amount != 4 || cached.Field547 != 2 || cached.Field546 != 11 {
					t.Fatal("NPC cached absorption/marker tail")
				}
				return true
			}
			if h, result := PlayerDamageNative4E17B0(v, a, a, 5, object.DamageImpact, r); !h || !result || exclusions != 1 || directions != 1 || wear != 1 || (replace && (live.Field547 != 55 || cached.Field1 != math.Float32bits(0.125))) {
				t.Fatal("NPC cached/live contract")
			}
		})
	}
}

func TestPlayerDamageNPCMonsterSelfImpactDefense4E17B0(t *testing.T) {
	for _, equipment := range []string{"shield", "sword", "staff"} {
		for _, front := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/front-%t", equipment, front), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
				ud := v.UpdateDataMonster()
				ud.Field547, ud.Field546 = 99, 77
				mask, sound := uint32(2), 878
				if equipment == "shield" {
					ud.ArmorEquipFlags = 0x1000000
					ud.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
				}
				if equipment == "sword" {
					ud.WeaponEquipFlags, mask, sound = 0x400, 0x400, 890
				}
				if equipment == "staff" {
					ud.WeaponEquipFlags, mask, sound = 0x8000, 0x8000, 894
				}
				item := &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 20, Max: 20}}
				v.InvFirstItem = item
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.BlockDirection = func(*Object, types.Pointf) bool { return front }
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
				wear := func(selected, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
					if selected != item || owner != v || attacker != a || attack != a || amount != 10 || typ != object.DamageImpact || ud.Field547 != 0 {
						t.Fatal("NPC block durability identity")
					}
					events = append(events, "wear")
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
				if front {
					r.DefaultDamage = nil
				} else {
					r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
						if source != a || weapon != a || amount != 40 || ud.Field547 != 2 || ud.Field546 != 11 {
							t.Fatal("NPC unblocked tail")
						}
						events = append(events, "default")
						return true
					}
				}
				if h, result := PlayerDamageNative4E17B0(v, a, a, 40, object.DamageImpact, r); !h || result == front {
					t.Fatalf("NPC defense=%t/%t", h, result)
				}
				want := []string{"default"}
				if front {
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

func TestPlayerDamageNPCMonsterSelfImpactActualHP4E17B0(t *testing.T) {
	for _, absorption := range []float32{0, 0.25} {
		for _, raw := range []int32{5, 40} {
			t.Run(fmt.Sprintf("armor-%g/raw-%d", absorption, raw), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "NPC"), monsterImpactTarget4E17B0(t, "monster")
				v.UpdateDataMonster().Field518 = math.Float32bits(absorption)
				r := damageMeleeRuntimeFixture4E17B0(t)
				effective := playerDamageRound4E17B0(float32((1 - float64(absorption)) * float64(raw)))
				if h, result := PlayerDamageNative4E17B0(v, a, a, raw, object.DamageImpact, r); !h || !result {
					t.Fatal("NPC actual HP path rejected")
				}
				ud := v.UpdateDataMonster()
				if v.HealthData.Cur != 200-uint16(effective) || ud.Field547 != 2 || ud.Field546 != 11 || !ud.StatusFlags.Has(object.MonStatusInjured) || v.Obj130 != a || v.Frame134 != 1400 || a.UpdateDataMonster().Field130 != 1400 {
					t.Fatal("NPC HP/Injured/attribution/source combat latch")
				}
			})
		}
	}
}
