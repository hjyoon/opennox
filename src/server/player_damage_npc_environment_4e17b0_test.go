package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// The stock jump table at 004E20A8 sends IMPALE (3) to 004E1F84's
// absorption/carry tail, but LAVA (12) to 004E1DC7's raw armor-wear tail.
func TestPlayerDamageNPCEnvironment4E17B0SignedArmorQuest(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageLava} {
		for _, raw := range []int32{-9, -1, 0, 1, 3, 9, 19} {
			for _, absorption := range []float32{0, .5, 1} {
				for _, initialCarry := range []float32{-.5, .25, .5} {
					for _, scale := range []float32{-1, 0, .5, 1.5} {
						t.Run(fmt.Sprintf("type-%d/raw-%d/armor-%g/carry-%g/quest-%g", typ, raw, absorption, initialCarry, scale), func(t *testing.T) {
							v, source, w := worldFlameFixture4E17B0(t, "self", "NPC")
							w.ObjClass = object.ClassDangerous | object.ClassImmobile
							if typ == object.DamageLava {
								source, w = nil, nil
								v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
							}
							armor := worldFlameArmorFixture4E17B0(t, v)
							if absorption == 0 {
								v.InvFirstItem = nil
							}
							ud := v.UpdateDataMonster()
							ud.Field518, ud.Field1 = math.Float32bits(absorption), math.Float32bits(initialCarry)
							ud.Field547, ud.Field546 = 99, 77
							want, carry, wear := raw, initialCarry, raw
							if typ == object.DamageImpale {
								accumulated := float32((1-float64(absorption))*float64(raw)) + initialCarry
								want = int32(math.RoundToEven(float64(accumulated)))
								carry = accumulated - float32(want)
								wear = raw - want
								if raw > 0 && want == 0 {
									want = 1
								}
							}
							if scale >= 0 {
								before := want
								want = int32(math.RoundToEven(float64(float32(float64(scale) * float64(want)))))
								if before > 0 && want < 1 {
									want = 1
								}
							}
							wears, defaults, gods, exclusions := 0, 0, 0, 0
							r := PlayerDamageRuntime4E17B0{
								BlockSourceExcluded: func(got *Object) bool {
									if got != w || typ == object.DamageLava || ud.Field547 != 0 {
										t.Fatal("NPC hazard exclusion prefix")
									}
									exclusions++
									return false
								},
								BlockDirection: func(got *Object, pos types.Pointf) bool {
									if got != v || w == nil || pos != w.PrevPos {
										t.Fatal("NPC hazard facing")
									}
									return true
								},
								ItemArmorValue: func(got *Object) float32 {
									if got != armor {
										t.Fatal("armor item")
									}
									return absorption
								},
								CanDamageArmor: func(got *Object) bool { return got == armor },
								DamageArmor: func(got, s, weapon *Object, amount int32, kind object.DamageType) bool {
									if got != armor || s != source || weapon != w || kind != typ || amount != wear || ud.Field1 != math.Float32bits(carry) {
										t.Fatalf("armor wear=%d/%d carry=%x/%x", amount, wear, ud.Field1, math.Float32bits(carry))
									}
									wears++
									return true
								},
								GodMode: func() bool {
									if ud.Field547 != 2 || ud.Field546 != uint32(typ) {
										t.Fatal("GodMode before cached hit marker")
									}
									gods++
									return true
								},
								QuestMode:        func() bool { return scale >= 0 },
								QuestDamageScale: func() float32 { return scale },
								DefaultDamage: func(got, s, weapon *Object, amount int32, kind object.DamageType) bool {
									if got != v || s != source || weapon != w || amount != want || kind != typ {
										t.Fatalf("HP tail=%d/%d type=%d", amount, want, kind)
									}
									defaults++
									return true
								},
								Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
							}
							h, result := PlayerDamageNative4E17B0(v, source, w, raw, typ, r)
							wantWear := 0
							if wear > 0 && absorption > 0 {
								wantWear = 1
							}
							wantExclusions := 1
							if typ == object.DamageLava {
								wantExclusions = 0
							}
							if !h || !result || defaults != 1 || gods != 1 || wears != wantWear || exclusions != wantExclusions || ud.Field1 != math.Float32bits(carry) {
								t.Fatalf("handled=%t result=%t defaults=%d gods=%d wear=%d/%d exclude=%d/%d", h, result, defaults, gods, wears, wantWear, exclusions, wantExclusions)
							}
						})
					}
				}
			}
		}
	}
}

func TestPlayerDamageNPCEnvironment4E17B0ShieldAndSnapshot(t *testing.T) {
	for _, mode := range []string{"front", "rear", "excluded", "lava"} {
		t.Run(mode, func(t *testing.T) {
			v, source, w := worldFlameFixture4E17B0(t, "NPC", "NPC")
			v.HealthData.Cur, v.HealthData.Max, v.HealthData.Field2 = 200, 200, 200
			w.ObjClass = object.ClassDangerous | object.ClassSimple
			ud := v.UpdateDataMonster()
			ud.ArmorEquipFlags, ud.Field547, ud.Field546 = 0x3000000, 99, 77
			ud.AIStackInd, ud.AIStack[0].Action = 0, uint32(ai.ACTION_BLOCK_ATTACK)
			typ, pos := object.DamageImpale, w.PrevPos
			if mode == "lava" {
				typ, source, w = object.DamageLava, nil, nil
			}
			r := damageFlameRuntime4E17B0(t, 0)
			exclusions, faces, sounds, wears := 0, 0, 0, 0
			r.BlockSourceExcluded = func(got *Object) bool {
				if got != w || ud.Field547 != 0 {
					t.Fatal("NPC spike exclusion order")
				}
				exclusions++
				w.PrevPos = types.Ptf(100, 200)
				return mode == "excluded"
			}
			r.BlockDirection = func(got *Object, at types.Pointf) bool {
				if got != v || at != pos || ud.Field547 != 1 || ud.Field546 != uint32(w.TypeInd) {
					t.Fatal("cached NPC attribution or position snapshot")
				}
				faces++
				return mode == "front"
			}
			r.Audio = func(id int, got *Object) {
				if id != 878 || got != v {
					t.Fatal("NPC shield sound")
				}
				sounds++
			}
			r.BlockDamagePercent = func() float64 { return .25 }
			r.DamageBlockItem = func(shield, got, s, weapon *Object, amount float32, kind object.DamageType) bool {
				if shield != nil || got != v || s != source || weapon != w || amount != 2.25 || kind != typ {
					t.Fatal("NPC live shield wear")
				}
				wears++
				return true
			}
			r.ProjectileReflect = func(*Object, *Object) { t.Fatal("reflected a spike") }
			h, result := PlayerDamageNative4E17B0(v, source, w, 9, typ, r)
			if !h {
				t.Fatal("NPC environmental prefix rejected")
			}
			if mode == "front" {
				if result || v.HealthData.Cur != 200 || sounds != 1 || wears != 1 || faces != 1 || exclusions != 1 {
					t.Fatalf("frontal NPC spike block: result=%t HP=%d sounds=%d wears=%d faces=%d exclusions=%d", result, v.HealthData.Cur, sounds, wears, faces, exclusions)
				}
			} else {
				wantFaces, wantExclusions := 1, 1
				if mode == "excluded" {
					wantFaces = 0
				}
				if mode == "lava" {
					wantFaces, wantExclusions = 0, 0
				}
				if !result || v.HealthData.Cur != 191 || sounds != 0 || wears != 0 || faces != wantFaces || exclusions != wantExclusions {
					t.Fatalf("NPC rear/excluded/lava health path: result=%t HP=%d sounds=%d wears=%d faces=%d/%d exclusions=%d/%d", result, v.HealthData.Cur, sounds, wears, faces, wantFaces, exclusions, wantExclusions)
				}
			}
		})
	}
}
