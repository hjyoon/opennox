package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestPlayerDamageSimpleCrushNative4E17B0ArmorMatrix(t *testing.T) {
	for _, fromPlayer := range []bool{false, true} {
		for _, toPlayer := range []bool{false, true} {
			for i, name := range []string{"SmallFist", "MediumFist", "LargeFist"} {
				t.Run(fmt.Sprintf("player-source-%t/player-target-%t/%s", fromPlayer, toPlayer, name), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, toPlayer)
					source := damageMeleeUnitFixture4E17B0(t, fromPlayer)
					fist := &Object{TypeInd: uint16(777 + i), ObjClass: object.ClassSimple, ObjOwner: source}
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					r := damageMeleeRuntimeFixture4E17B0(t)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					reason := ""
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageMeleeNative4E17B0(target, source, fist, 9, object.DamageCrush, r)
					marker, markerType, carry := damageMeleeMarker4E17B0(target)
					// CRUSH absorbs only half the armor coefficient: 9*0.75+0.4
					// rounds to 7, retains 0.15 carry, and wears armor by 2.
					if !h || !result || target.HealthData.Cur != 193 || armor.HealthData.Cur != 23 ||
						marker != 1 || markerType != uint32(fist.TypeInd) || math.Abs(float64(carry-0.15)) > 1e-6 || reason != "" {
						t.Fatalf("Fist defense: handled=%t/%t HP=200->%d armor=%d marker=%d/%d carry=%g reason=%q",
							h, result, target.HealthData.Cur, armor.HealthData.Cur, marker, markerType, carry, reason)
					}
				})
			}
		}
	}
}

func TestPlayerDamageSimpleCrushNative4E17B0OriginalBlockRules(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, excluded := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-%t/excluded-%t", player, excluded), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, player)
				source := damageMeleeUnitFixture4E17B0(t, !player)
				fist := &Object{TypeInd: 777, ObjClass: object.ClassSimple, ObjOwner: source}
				target.Buffs = 1 << playerDamageReflectEnchant4E17B0
				if player {
					ud := target.UpdateDataPlayer()
					ud.State, ud.Player.ArmorEquip = PlayerState16, 0x1000000
				} else {
					ud := target.UpdateDataMonster()
					ud.ArmorEquipFlags, ud.AIStack[0].Action = 0x1000000, uint32(ai.ACTION_BLOCK_ATTACK)
				}
				shield := &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}
				target.InvFirstItem = shield
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.BlockSourceExcluded = func(got *Object) bool {
					if got != fist {
						t.Fatal("block exclusion did not inspect the Fist")
					}
					return excluded
				}
				r.BlockDirection = func(*Object, types.Pointf) bool {
					if excluded {
						t.Fatal("stock Fist must bypass ordinary shield direction/durability")
					}
					return true
				}
				r.ProjectileReflect = func(*Object, *Object) { t.Fatal("SIMPLE CRUSH is not a missile") }
				r.Audio = func(id int, _ *Object) {
					if id != 878 || excluded {
						t.Fatal("unexpected Fist block audio")
					}
				}
				r.BlockDamagePercent = func() float64 { return 0.2 }
				r.CanDamageBlockItem = func(item *Object) bool { return item == shield }
				r.DamageBlockItem = func(item, owner, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
					if excluded || item != shield || owner != target || attacker != source || weapon != fist || amount != 2 || typ != object.DamageCrush {
						t.Fatal("non-excluded SIMPLE shield durability arguments")
					}
					return true
				}
				r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("shield-only hit must not change state"); return false }
				r.Melee.MonsterPopBlockAction = func(*Object) { t.Fatal("intact shield must not pop action") }
				h, result := PlayerDamageMeleeNative4E17B0(target, source, fist, 10, object.DamageCrush, r)
				want := uint16(200)
				if excluded {
					want = 190
				}
				if !h || result != excluded || target.HealthData.Cur != want {
					t.Fatalf("SIMPLE block: handled=%t result=%t HP=%d want=%d", h, result, target.HealthData.Cur, want)
				}
			})
		}
	}
}

func TestPlayerDamageSimpleCrushNative4E17B0GodModeQuestAndCoop(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, mode := range []string{"god", "quest", "coop-owned"} {
			t.Run(fmt.Sprintf("player-%t/%s", player, mode), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, player)
				source := damageMeleeUnitFixture4E17B0(t, !player)
				fist := &Object{ObjClass: object.ClassSimple, ObjOwner: source}
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				wantHP, wantArmor, wantResult := uint16(193), uint16(23), true
				switch mode {
				case "god":
					r.GodMode = func() bool { return true }
					if player {
						wantHP = 200
					} // Armor durability still precedes GodMode.
				case "quest":
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = func() float32 { return 0.5 }
					wantHP = 196 // 7*0.5 rounds to even 4, after armor/carry.
				case "coop-owned":
					source.ObjOwner = target
					r.CoopMode = func() bool { return true }
					if player {
						wantHP, wantArmor, wantResult = 200, 25, false
					}
				}
				if h, result := PlayerDamageMeleeNative4E17B0(target, source, fist, 9, object.DamageCrush, r); !h || result != wantResult ||
					target.HealthData.Cur != wantHP || armor.HealthData.Cur != wantArmor {
					t.Fatalf("Fist %s: HP=%d/%d armor=%d/%d handled=%t result=%t/%t", mode, target.HealthData.Cur, wantHP, armor.HealthData.Cur, wantArmor, h, result, wantResult)
				}
			})
		}
	}
}
