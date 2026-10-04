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

// Record boundary arguments for signed/zero damage instead of manufacturing
// a negative-HP result. Separate positive cases use the real native Go tail.
func TestPlayerDamageMonsterSelfImpactArmor4E17B0(t *testing.T) {
	for _, absorption := range []float32{0, 0.25, 1} {
		for _, raw := range []int32{-3, 0, 1, 5, 40} {
			for _, mode := range []string{"normal", "quest", "god"} {
				t.Run(fmt.Sprintf("armor-%g/raw-%d/%s", absorption, raw, mode), func(t *testing.T) {
					v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
					ud := v.UpdateDataPlayer()
					ud.Field57, ud.Field21 = math.Float32bits(absorption), math.Float32bits(0.125)
					ud.Field76, ud.Field75 = 99, 77
					r := damageMeleeRuntimeFixture4E17B0(t)
					var events []string
					r.BlockSourceExcluded = func(attack *Object) bool {
						if attack != a || ud.Field76 != 0 || ud.Field75 != 77 {
							t.Fatal("self-weapon entry marker")
						}
						events = append(events, "exclude")
						return false
					}
					r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "facing"); return false }
					accumulated := float32((1-float64(absorption))*float64(raw)) + 0.125
					unscaled := playerDamageRound4E17B0(accumulated)
					effective := unscaled
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
						if ud.Field76 != 2 || ud.Field75 != 11 || ud.Field21 != math.Float32bits(accumulated-float32(unscaled)) {
							t.Fatal("God preceded carry/type marker")
						}
						events = append(events, "god")
						return mode == "god"
					}
					r.QuestMode = func() bool { events = append(events, "quest"); return mode == "quest" }
					r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
					r.DefaultDamage = func(target, source, weapon *Object, amount int32, typ object.DamageType) bool {
						if target != v || source != a || weapon != a || amount != effective || typ != object.DamageImpact {
							t.Fatalf("default arguments: %d want %d", amount, effective)
						}
						events = append(events, "default")
						return true
					}
					if h, result := PlayerDamageNative4E17B0(v, a, a, raw, object.DamageImpact, r); !h || !result {
						t.Fatal("self-weapon IMPACT rejected")
					}
					want := []string{"exclude", "facing", "god"}
					if mode != "god" {
						want = append(want, "quest")
						if mode == "quest" {
							want = append(want, "scale")
						}
						want = append(want, "default")
					}
					if !slices.Equal(events, want) || v.HealthData.Cur != 200 {
						t.Fatalf("events=%v want=%v", events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamageMonsterSelfImpactCachedPrefix4E17B0(t *testing.T) {
	for _, observe := range []bool{false, true} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("observe-%t/replace-%t", observe, replace), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
				cached := v.UpdateDataPlayer()
				cached.Field57, cached.Field21 = math.Float32bits(0.25), math.Float32bits(0.125)
				cached.Field76, cached.Field75 = 99, 77
				if observe {
					cached.Player.Field3680, cached.Player.CameraFollowObj = 2, a
				}
				live := cached
				armor := damageMeleeArmorFixture4E17B0(v, 0.25, 0.125)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
				var events []string
				r.ObserveClear = func(*Object) {
					if !observe || cached.Field76 != 0 {
						t.Fatal("observe before cleared marker")
					}
					cached.Player.CameraFollowObj = nil
					events = append(events, "observe")
				}
				pos := a.PrevPos
				r.BlockSourceExcluded = func(*Object) bool {
					if cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("self-weapon was attributed as distinct weapon")
					}
					if replace {
						live = &PlayerUpdateData{Player: &Player{}, Field76: 55}
						v.UpdateData = unsafe.Pointer(live)
					}
					live.Field57, live.Field21 = math.Float32bits(0.875), math.Float32bits(0.5)
					a.PrevPos = types.Ptf(-100, 99)
					events = append(events, "exclude")
					return false
				}
				r.BlockDirection = func(_ *Object, attackPos types.Pointf) bool {
					if attackPos != pos || cached.Field76 != 0 {
						t.Fatal("lost directed prefix snapshot")
					}
					events = append(events, "facing")
					return false
				}
				r.ItemArmorValue = func(*Object) float32 { events = append(events, "armor-value"); return 0.875 }
				r.DamageArmor = func(_, source, weapon *Object, amount int32, typ object.DamageType) bool {
					if source != a || weapon != a || amount != 1 || typ != object.DamageImpact || cached.Field76 != 0 || live.Field21 != math.Float32bits(0.25) {
						t.Fatal("full-armor carry/wear ordering")
					}
					events = append(events, "wear")
					return true
				}
				r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
					if source != a || weapon != a || amount != 4 || cached.Field76 != 2 || cached.Field75 != 11 {
						t.Fatal("entry armor/type fallback changed")
					}
					events = append(events, "default")
					return true
				}
				if h, result := PlayerDamageNative4E17B0(v, a, a, 5, object.DamageImpact, r); !h || !result {
					t.Fatal("cached prefix rejected")
				}
				want := []string{"exclude", "facing", "armor-value", "wear", "default"}
				if observe {
					want = append([]string{"observe"}, want...)
				}
				if !slices.Equal(events, want) || (replace && (cached.Field21 != math.Float32bits(0.125) || live.Field76 != 55)) {
					t.Fatalf("events=%v carry cached=%g live=%g", events, math.Float32frombits(cached.Field21), math.Float32frombits(live.Field21))
				}
			})
		}
	}
}

func TestPlayerDamageMonsterSelfImpactDefense4E17B0(t *testing.T) {
	for _, equipment := range []string{"shield", "sword", "staff"} {
		for _, front := range []bool{false, true} {
			for _, idle := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/front-%t/idle-%t", equipment, front, idle), func(t *testing.T) {
					v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
					ud := v.UpdateDataPlayer()
					ud.Field76, ud.Field75 = 99, 77
					mask, sound := uint32(2), 878
					if idle {
						ud.State = PlayerState13
					} else {
						ud.State = PlayerState1
					}
					switch equipment {
					case "shield":
						ud.Player.ArmorEquip, ud.State = 0x1000000, PlayerState16
					case "sword":
						ud.Player.WeaponEquip, mask, sound = 0x400, 0x400, 890
					case "staff":
						ud.Player.WeaponEquip, mask, sound = 0x8000, 0x8000, 894
					}
					item := &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 20, Max: 20}}
					v.InvFirstItem = item
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.BlockDirection = func(*Object, types.Pointf) bool { return front }
					blocked := front && (equipment == "shield" || idle)
					var events []string
					r.Audio = func(id int, _ *Object) {
						if id != sound || ud.Field76 != 0 {
							t.Fatal("block audio/prefix")
						}
						events = append(events, "audio")
					}
					r.PlayerSetState = func(_ *Object, state PlayerState) bool {
						want := PlayerState21
						if equipment == "sword" {
							want = PlayerState19
						}
						if state != want {
							t.Fatal("block animation")
						}
						events = append(events, "state")
						return true
					}
					r.Melee.RandomInt = func(lo, hi int) int {
						if lo != 18 || hi != 20 {
							t.Fatal("block range")
						}
						events = append(events, "random")
						return 19
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "percent"); return 0.25 }
					r.CanDamageBlockItem = func(*Object) bool { return true }
					r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
					wear := func(selected, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
						if selected != item || owner != v || attacker != a || attack != a || amount != 10 || typ != object.DamageImpact || ud.Field76 != 0 {
							t.Fatal("block durability identity/amount")
						}
						events = append(events, "wear")
						return true
					}
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
					if blocked {
						r.DefaultDamage = nil
					} else {
						r.DefaultDamage = func(_, source, weapon *Object, amount int32, _ object.DamageType) bool {
							if source != a || weapon != a || amount != 40 || ud.Field76 != 2 || ud.Field75 != 11 {
								t.Fatal("unblocked tail")
							}
							events = append(events, "default")
							return true
						}
					}
					if h, result := PlayerDamageNative4E17B0(v, a, a, 40, object.DamageImpact, r); !h || result == blocked {
						t.Fatalf("defense=%t/%t", h, result)
					}
					want := []string{"default"}
					if blocked {
						want = []string{"audio"}
						if equipment == "sword" {
							want = append(want, "random")
						}
						if equipment != "shield" {
							want = append(want, "state")
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

func TestPlayerDamageMonsterSelfImpactActualHP4E17B0(t *testing.T) {
	for _, absorption := range []float32{0, 0.25} {
		for _, raw := range []int32{5, 40} {
			t.Run(fmt.Sprintf("armor-%g/raw-%d", absorption, raw), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
				v.UpdateDataPlayer().Field57 = math.Float32bits(absorption)
				r := damageMeleeRuntimeFixture4E17B0(t)
				effective := playerDamageRound4E17B0(float32((1 - float64(absorption)) * float64(raw)))
				if h, result := PlayerDamageNative4E17B0(v, a, a, raw, object.DamageImpact, r); !h || !result {
					t.Fatal("actual HP path rejected")
				}
				ud := v.UpdateDataPlayer()
				if v.HealthData.Cur != 200-uint16(effective) || ud.Field76 != 2 || ud.Field75 != 11 || v.Obj130 != a || v.Frame134 != 1400 || a.UpdateDataMonster().Field130 != 1400 || (effective >= 20 && ud.State != PlayerState30) {
					t.Fatal("HP/attribution/hurt/source combat latch")
				}
			})
		}
	}
}
