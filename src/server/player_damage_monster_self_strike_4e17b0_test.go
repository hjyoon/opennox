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

// GAME.EXE 004E20A8 routes CRUSH through half armor and BLADE/IMPALE/CLAW
// through full armor. Signed and zero inputs still visit carry and markers.
func TestPlayerDamageMonsterSelfStrikeArmor4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageClaw} {
		for _, armor := range []float32{0, 0.25, 1} {
			for _, raw := range []int32{-3, 0, 1, 5, 40} {
				for _, mode := range []string{"normal", "quest", "god"} {
					t.Run(fmt.Sprintf("type-%d/armor-%g/raw-%d/%s", typ, armor, raw, mode), func(t *testing.T) {
						v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
						ud := v.UpdateDataPlayer()
						ud.Field57, ud.Field21 = math.Float32bits(armor), math.Float32bits(0.125)
						ud.Field76, ud.Field75 = 99, 77
						absorption := float64(armor)
						if typ == object.DamageCrush {
							absorption *= 0.5
						}
						accumulated := float32((1-absorption)*float64(raw)) + 0.125
						unscaled := int32(math.RoundToEven(float64(accumulated)))
						amount := unscaled
						if raw > 0 && amount == 0 {
							amount = 1
						}
						if mode == "quest" {
							before := amount
							amount = int32(math.RoundToEven(float64(float32(float64(amount) * 0.5))))
							if before > 0 && amount < 1 {
								amount = 1
							}
						}
						r := damageMeleeRuntimeFixture4E17B0(t)
						var events []string
						r.BlockSourceExcluded = func(attack *Object) bool {
							if attack != a || ud.Field76 != 0 || ud.Field75 != 77 {
								t.Fatal("self-weapon prefix marker")
							}
							events = append(events, "exclude")
							return false
						}
						r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "facing"); return false }
						r.GodMode = func() bool {
							if ud.Field76 != 2 || ud.Field75 != uint32(typ) || ud.Field21 != math.Float32bits(accumulated-float32(unscaled)) {
								t.Fatal("God before cached marker/live carry")
							}
							events = append(events, "god")
							return mode == "god"
						}
						r.QuestMode = func() bool { events = append(events, "quest"); return mode == "quest" }
						r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
						r.DefaultDamage = func(target, source, weapon *Object, damage int32, dt object.DamageType) bool {
							if target != v || source != a || weapon != a || damage != amount || dt != typ {
								t.Fatalf("native tail arguments: %d want %d", damage, amount)
							}
							events = append(events, "default")
							return true
						}
						if h, result := PlayerDamageNative4E17B0(v, a, a, raw, typ, r); !h || !result {
							t.Fatal("monster self-strike rejected")
						}
						want := []string{"exclude", "facing", "god"}
						if mode != "god" {
							want = append(want, "quest")
							if mode == "quest" {
								want = append(want, "scale")
							}
							want = append(want, "default")
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

func TestPlayerDamageMonsterSelfStrikeCachedPrefix4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageCrush, object.DamageImpale, object.DamageClaw} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
			cached := v.UpdateDataPlayer()
			armor := damageMeleeArmorFixture4E17B0(v, 0.25, 0.125)
			live := &PlayerUpdateData{Player: &Player{}, Field21: math.Float32bits(0.5), Field57: math.Float32bits(1), Field76: 55}
			cached.Field76, cached.Field75 = 99, 77
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
			pos := a.PrevPos
			r.BlockSourceExcluded = func(*Object) bool {
				if cached.Field76 != 0 {
					t.Fatal("entry marker not cleared")
				}
				v.UpdateData, a.PrevPos = unsafe.Pointer(live), types.Ptf(-100, 99)
				return false
			}
			r.BlockDirection = func(_ *Object, p types.Pointf) bool {
				if p != pos {
					t.Fatal("lost position snapshot")
				}
				return false
			}
			amount, carry := int32(4), float32(0.25)
			if typ == object.DamageCrush {
				amount, carry = 5, -0.125
			}
			r.DefaultDamage = func(_, source, weapon *Object, damage int32, dt object.DamageType) bool {
				if source != a || weapon != a || damage != amount || dt != typ || cached.Field76 != 2 || cached.Field75 != uint32(typ) || live.Field21 != math.Float32bits(carry) {
					t.Fatal("cached armor/marker vs live carry")
				}
				return true
			}
			if h, result := PlayerDamageNative4E17B0(v, a, a, 5, typ, r); !h || !result {
				t.Fatal("cached self-strike rejected")
			}
			if live.Field76 != 55 || cached.Field21 != math.Float32bits(0.125) || armor.HealthData.Cur != uint16(25-(5-amount)) {
				t.Fatal("wrong record mutated")
			}
		})
	}
}

func TestPlayerDamageMonsterSelfStrikeRealHPTail4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageBlade, object.DamageCrush, object.DamageImpale, object.DamageDrain, object.DamageClaw} {
		v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
		r := damageMeleeRuntimeFixture4E17B0(t)
		if h, result := PlayerDamageNative4E17B0(v, a, a, 10, typ, r); !h || !result || v.HealthData.Cur != 190 || v.Obj130 != a || v.Field131 != uint32(typ) || a.UpdateDataMonster().Field130 != 1400 {
			t.Fatalf("real native HP tail rejected type %d: HP=%d", typ, v.HealthData.Cur)
		}
	}
}

func TestPlayerDamageMonsterSelfStrikeDrainRaw4E17B0(t *testing.T) {
	for _, raw := range []int32{-3, 0, 1, 5, 40} {
		for _, mode := range []string{"normal", "quest", "god"} {
			t.Run(fmt.Sprintf("raw-%d/%s", raw, mode), func(t *testing.T) {
				v, a := monsterImpactTarget4E17B0(t, "player"), monsterImpactTarget4E17B0(t, "monster")
				ud := v.UpdateDataPlayer()
				armor := damageMeleeArmorFixture4E17B0(v, 1, 0.125)
				ud.Field76, ud.Field75 = 99, 77
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.BlockDirection = func(*Object, types.Pointf) bool { return false }
				r.GodMode = func() bool { return mode == "god" }
				r.QuestMode = func() bool { return mode == "quest" }
				r.QuestDamageScale = func() float32 { return 0.5 }
				called := false
				r.DefaultDamage = func(target, source, weapon *Object, d int32, typ object.DamageType) bool {
					want := raw
					if mode == "quest" {
						want = int32(math.RoundToEven(float64(float32(float64(raw) * 0.5))))
						if raw > 0 && want < 1 {
							want = 1
						}
					}
					if target != v || source != a || weapon != a || typ != object.DamageDrain || d != want {
						t.Fatalf("DRAIN default damage = %d, want %d", d, want)
					}
					called = true
					return true
				}
				if h, result := PlayerDamageNative4E17B0(v, a, a, raw, object.DamageDrain, r); !h || !result {
					t.Fatal("DRAIN rejected")
				}
				if called != (mode != "god") || ud.Field21 != math.Float32bits(0.125) || armor.HealthData.Cur != 25 || ud.Field76 != 2 || ud.Field75 != 4 {
					t.Fatal("DRAIN changed armor/carry or skipped original marker/God ordering")
				}
			})
		}
	}
}
