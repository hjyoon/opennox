package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

func TestDefaultDamageEnvironmentProtectionAndSignedTail(t *testing.T) {
	for _, victim := range []string{"monster", "NPC"} {
		for _, hazard := range []string{"lava", "spike"} {
			for _, raw := range []int32{-3, 0, 1, 3, 9} {
				for _, immune := range []bool{false, true} {
					for _, protection := range []float64{0, .5, 1} {
						t.Run(fmt.Sprintf("%s/%s/raw-%d/immune-%t/protection-%g", victim, hazard, raw, immune, protection), func(t *testing.T) {
							v, _, w := worldFlameFixture4E17B0(t, "self", victim)
							var source, weapon *Object
							typ := object.DamageLava
							if hazard == "spike" {
								typ = object.DamageImpale
								w.ObjClass = object.ClassDangerous | object.ClassImmobile
								source, weapon = w, w
							}
							if immune {
								v.ObjSubClass |= 0x400
							}
							blocked := hazard == "lava" && immune
							want := raw
							if hazard == "lava" {
								want = int32(math.RoundToEven(float64(float32((1 - float64(float32(protection))) * float64(raw)))))
								if want == 0 {
									want = 1
								}
							}
							calls, fire, invis := 0, 0, 0
							r := DefaultDamageWorldRuntime4E0B30{
								Frame: func() uint32 { return 1400 }, GameplayFlag1: func() bool { return true },
								IsEnemy:            func(*Object, *Object) bool { return false },
								FireProtection:     func(*Object) float64 { fire++; return protection },
								ElectricProtection: func(*Object) float64 { t.Fatal("environment used electric resistance"); return 0 },
								BuffOff: func(got *Object, enchant EnchantID) {
									if got != v || enchant != 0 {
										t.Fatal("environment invisibility service")
									}
									invis++
								},
								DamageClear: func(got *Object, amount int32) {
									if got != v || amount != want {
										t.Fatalf("HP amount=%d want=%d", amount, want)
									}
									calls++
								},
								Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
							}
							if !DefaultDamageWorld4E0B30(v, source, weapon, raw, typ, r) {
								t.Fatal("environment return")
							}
							wantCalls := 1
							if blocked {
								wantCalls = 0
							}
							wantFire := 0
							if hazard == "lava" && !blocked {
								wantFire = 1
							}
							if calls != wantCalls || invis != wantCalls || fire != wantFire {
								t.Fatalf("HP=%d invis=%d fire=%d blocked=%t", calls, invis, fire, blocked)
							}
							ud := v.UpdateDataMonster()
							if ud.Field1 != math.Float32bits(.25) {
								t.Fatal("DefaultDamage changed armor carry")
							}
							if !blocked && (ud.Field547 != 2 || ud.Field546 != uint32(typ) || ud.StatusFlags&object.MonStatusInjured == 0 || v.Obj130 != weapon || v.Field131 != uint32(typ)) {
								t.Fatal("environment attribution/injured marker")
							}
							if hazard == "lava" && ud.StatusFlags&object.MonStatusOnFire == 0 {
								t.Fatal("lava entry OnFire flag")
							}
						})
					}
				}
			}
		}
	}
}
