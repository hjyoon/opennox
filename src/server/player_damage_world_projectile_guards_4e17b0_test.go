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

func TestPlayerDamageWorldProjectile4E17B0SignedCasesAndLateQuest(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, tc := range worldProjectileCases4E0B30() {
			for _, damage := range []int32{-9, 0, 1, 9} {
				t.Run(fmt.Sprintf("player-%t/%s/%d", player, tc.name, damage), func(t *testing.T) {
					v := damageMeleeUnitFixture4E17B0(t, player)
					a, w, _ := worldProjectileFixture4E0B30(tc)
					armor := damageMeleeArmorFixture4E17B0(v, 0.5, 0.4)
					r := damageMeleeRuntimeFixture4E17B0(t)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					quest := false
					wear := int32(123456)
					got := int32(123456)
					r.DamageArmor = func(_, _, _ *Object, d int32, k object.DamageType) bool { wear = d; quest = true; return true }
					r.QuestMode = func() bool { return quest }
					r.QuestDamageScale = func() float32 { return 0.5 }
					r.DefaultDamage = func(_, _, _ *Object, d int32, k object.DamageType) bool { got = d; return true }
					before := damage
					wantCarry := float32(0.4)
					if tc.typ != object.DamageFlame {
						scale := 0.5
						if tc.typ == object.DamageCrush {
							scale = 0.25
						}
						acc := float32((1-scale)*float64(damage)) + float32(0.4)
						before = playerDamageRound4E17B0(acc)
						wantCarry = acc - float32(before)
					}
					wantWear := damage - before
					if tc.typ == object.DamageFlame {
						wantWear = damage
					}
					want := before
					if damage > 0 && want == 0 {
						want = 1
					}
					if wantWear > 0 {
						prior := want
						want = playerDamageRound4E17B0(float32(float64(want) * 0.5))
						if prior > 0 && want < 1 {
							want = 1
						}
					}
					if h, ok := PlayerDamageNative4E17B0(v, a, w, damage, tc.typ, r); !h || !ok {
						t.Fatal("signed projectile rejected")
					}
					_, _, carry := damageMeleeMarker4E17B0(v)
					if got != want || carry != wantCarry || (wantWear > 0 && wear != wantWear) {
						t.Fatalf("HP input=%d/%d wear=%d/%d carry=%g/%g", got, want, wear, wantWear, carry, wantCarry)
					}
				})
			}
		}
	}
}

func TestPlayerDamageWorldProjectile4E17B0CachedMarkerLiveCarry(t *testing.T) {
	for _, player := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-%t", player), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, player)
			a, w, m := worldProjectileFixture4E0B30(worldProjectileCases4E0B30()[0])
			var marker, kind *uint32
			if player {
				u := v.UpdateDataPlayer()
				u.Field57 = math.Float32bits(0.5)
				marker, kind = &u.Field76, &u.Field75
			} else {
				u := v.UpdateDataMonster()
				u.Field518 = math.Float32bits(0.5)
				marker, kind = &u.Field547, &u.Field546
			}
			*marker, *kind = 77, 88
			r := damageMeleeRuntimeFixture4E17B0(t)
			pos := m.PrevPos
			queries := 0
			r.BlockSourceExcluded = func(*Object) bool {
				if *marker != 0 {
					t.Fatal("prefix marker not cleared")
				}
				if player {
					v.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &Player{}, Field21: math.Float32bits(0.75), Field76: 66})
				} else {
					v.UpdateData = unsafe.Pointer(&MonsterUpdateData{Field1: math.Float32bits(0.75), Field547: 66})
				}
				m.PrevPos = types.Ptf(999, 888)
				m.TypeInd = 222
				return false
			}
			r.BlockDirection = func(_ *Object, p types.Pointf) bool {
				queries++
				if p != pos || *marker != 1 || *kind != 222 {
					t.Fatal("position/marker not retained at facing")
				}
				return false
			}
			r.DefaultDamage = func(_, _, _ *Object, d int32, k object.DamageType) bool {
				if d != 5 {
					t.Fatalf("damage=%d", d)
				}
				return true
			}
			if h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageImpale, r); !h || !ok {
				t.Fatal("replaced record rejected")
			}
			liveMarker, _, carry := damageMeleeMarker4E17B0(v)
			if queries != 1 || *marker != 1 || *kind != 222 || liveMarker != 66 || carry != 0.25 {
				t.Fatalf("cached/live marker=%d/%d carry=%g", *marker, liveMarker, carry)
			}
		})
	}
}

func TestPlayerDamageWorldProjectile4E17B0Defenses(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, defense := range []string{"Reflect", "Shield", "GreatSword"} {
			for _, retain := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-%t/%s/bit2-%t", player, defense, retain), func(t *testing.T) {
					v := damageMeleeUnitFixture4E17B0(t, player)
					a, w, m := worldProjectileFixture4E0B30(worldProjectileCases4E0B30()[3])
					m.ObjSubClass = 0
					if retain {
						m.ObjSubClass = 2
					}
					item := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 25, Max: 25}}
					v.InvFirstItem = item
					if defense == "Reflect" {
						v.Buffs = 1 << playerDamageReflectEnchant4E17B0
					}
					if defense == "Shield" {
						if player {
							u := v.UpdateDataPlayer()
							u.Player.ArmorEquip = 0x1000000
							u.State = PlayerState16
						} else {
							u := v.UpdateDataMonster()
							u.ArmorEquipFlags = 0x1000000
							u.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
						}
					}
					if defense == "GreatSword" {
						item.ObjClass = object.ClassWeapon
						item.ObjSubClass = 0x400
						if player {
							v.UpdateDataPlayer().Player.WeaponEquip = 0x400
						} else {
							u := v.UpdateDataMonster()
							u.WeaponEquipFlags = 0x400
							u.AIStack[0].Action = uint32(ai.ACTION_GUARD)
						}
					}
					r := damageMeleeRuntimeFixture4E17B0(t)
					var events []string
					directions, exclusions := 0, 0
					r.BlockSourceExcluded = func(*Object) bool { exclusions++; return false }
					r.BlockDirection = func(*Object, types.Pointf) bool { directions++; return true }
					r.ProjectileReflect = func(got, owner *Object) {
						marker, _, _ := damageMeleeMarker4E17B0(v)
						want := uint32(1)
						if defense == "Reflect" {
							want = 0
						}
						if got != m || owner != v || marker != want {
							t.Fatal("reflection prefix lost")
						}
						events = append(events, "reflect")
					}
					r.ClearOwner = func(o *Object) { events = append(events, "clear"); o.ObjOwner = nil }
					r.SetOwner = func(owner, o *Object) { events = append(events, "set"); o.ObjOwner = owner }
					r.ChangeOwner = func(o, owner *Object) { events = append(events, "change") }
					r.Audio = func(n int, _ *Object) { events = append(events, fmt.Sprintf("audio-%d", n)) }
					r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.2 }
					r.CanDamageBlockItem = func(*Object) bool { return true }
					r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
					wear := func(i, owner, source, effective *Object, d float32, k object.DamageType) bool {
						if i != item || owner != v || source != a || effective != w || d != float32(1.8) || k != object.DamageCrush {
							t.Fatal("block wear inputs lost")
						}
						events = append(events, "wear")
						return true
					}
					r.DamageBlockItem = wear
					r.Melee.DamageBlockWeapon = wear
					r.Melee.RandomInt = func(int, int) int { return 19 }
					r.PlayerSetState = func(_ *Object, s PlayerState) bool { events = append(events, "state"); return true }
					r.Melee.MonsterBlockAction = func(*Object) { events = append(events, "state") }
					r.Melee.MonsterPopBlockAction = func(*Object) {}
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("blocked projectile reached HP")
						return false
					}
					if h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageCrush, r); !h || ok {
						t.Fatalf("defense=%t/%t", h, ok)
					}
					want := []string{}
					if defense == "Shield" {
						want = append(want, "audio-878")
					}
					want = append(want, "reflect")
					owner := a
					if !retain || defense == "Reflect" {
						want = append(want, "clear", "set")
						owner = v
					}
					if defense == "Reflect" {
						if retain {
							want = append(want, "change")
						}
						want = append(want, "audio-122")
					} else {
						if defense == "GreatSword" {
							want = append(want, "audio-890", "state")
						}
						want = append(want, "balance", "wear")
					}
					wantExclusions := 1
					if defense == "Reflect" {
						wantExclusions = 0
					}
					if !slices.Equal(events, want) || directions != 1 || exclusions != wantExclusions || m.ObjOwner != owner || v.HealthData.Cur != 200 {
						t.Fatalf("events=%v want=%v directions=%d exclusions=%d", events, want, directions, exclusions)
					}
				})
			}
		}
	}
}
