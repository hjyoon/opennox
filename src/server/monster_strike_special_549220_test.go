package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// The PE pushes literal actions 25, 41, 24, not MOVE_TO (7). Exercise the
// real native stack, including partial capacity, and keep the original RNG
// draw conditional on accepting the TIME push.
func TestMonsterStrikeSpecial549220GhostNativeStack(t *testing.T) {
	for slots := 0; slots <= 3; slots++ {
		for _, fps := range []uint32{30, 0x80000001, 0xffffffff} {
			t.Run(fmt.Sprintf("slots-%d/fps-%08x", slots, fps), func(t *testing.T) {
				_, unit, update := moveToNativeRetryFixture5443F0(t, ai.ACTION_MELEE_ATTACK)
				update.MonsterDef = &MonsterDef{MeleeAttackDamage116: 10, MeleeAttackDamageType124: 4, MeleeAttackImpact120: 2}
				update.AIStackInd = int8(len(update.AIStack) - 1 - slots)
				update.AIStackHead().Action = 16
				base := int(update.AIStackInd)
				target := &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(120, 210)}
				frame, randomCalls := uint32(1000), 0
				var events []string
				h := monsterStrikeSpecialHooks549220{
					MonsterStrikeSpecialRuntime549220: MonsterStrikeSpecialRuntime549220{
						MonsterStrikeDefaultRuntime549380: MonsterStrikeDefaultRuntime549380{
							Damage: func(v, source, weapon *Object, damage int, typ object.DamageType) bool {
								if v != target || source != unit || weapon != unit || damage != 10 || typ != 4 {
									t.Fatal("native drain identity")
								}
								events = append(events, "damage")
								return true
							},
							ApplyForce: func(v *Object, pos types.Pointf, force float64) {
								if v != target || pos != unit.PosVec || force != 2 {
									t.Fatal("ghost force")
								}
								events = append(events, "force")
							},
						},
						BuffApply: func(v *Object, enchant EnchantID, duration, power int) {
							if v != target || enchant != 5 || duration != int(2*fps) || power != 3 {
								t.Fatal("ghost buff DWORD duration")
							}
							target.PosVec = types.Ptf(130, 220)
							events = append(events, "buff")
						},
					},
					pick: func(v *Object) *Object {
						if v != unit {
							t.Fatal("native picker")
						}
						return target
					},
					trace:    func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
					tickRate: func() uint32 { events = append(events, "fps"); return fps },
					random: func(lo, hi int) int {
						if lo != int(int32(2*fps)) || hi != int(int32(4*fps)) {
							t.Fatal("signed DWORD random bounds")
						}
						randomCalls++
						frame = 0xfffffffe
						events = append(events, "random")
						return lo
					},
					frame: func() uint32 { events = append(events, "frame"); return frame },
				}
				if monsterStrikeSpecial549220(unit, MonsterStrikeGhost549220, h) != 1 || int(update.AIStackInd) != base+slots {
					t.Fatal("native ghost result/capacity")
				}
				want := []string{"damage", "force", "fps", "buff"}
				if slots >= 2 {
					want = append(want, "fps", "random", "frame")
				}
				assertMeleeEvents549380(t, events, want)
				wantRandom := 0
				if slots >= 2 {
					wantRandom = 1
				}
				if randomCalls != wantRandom {
					t.Fatal("RNG on rejected TIME push")
				}
				for i, action := range []uint32{25, 41, 24}[:slots] {
					item := update.AIStack[base+1+i]
					var args [4]uintptr
					if action == 41 {
						args[0] = uintptr(uint32(0xfffffffe) + 2*fps)
					} else {
						args[0], args[1] = uintptr(math.Float32bits(130)), uintptr(math.Float32bits(220))
					}
					if item.Action != action || item.Args != args || item.Field5 != 0 {
						t.Fatalf("stack=%+v want literal action=%d args=%v", item, action, args)
					}
				}
			})
		}
	}
}

func TestMonsterStrikeSpecial549220SingleMissResults(t *testing.T) {
	for _, kind := range []MonsterStrikeKind549220{MonsterStrikeScorpion549220, MonsterStrikeZombie549220, MonsterStrikeWasp549220, MonsterStrikeGhost549220} {
		for _, found := range []bool{false, true} {
			t.Run(fmt.Sprintf("kind-%d/found-%t", kind, found), func(t *testing.T) {
				unit := meleeMonsterTestObject532130(t)
				target := &Object{}
				traces := 0
				result := monsterStrikeSpecial549220(unit, kind, monsterStrikeSpecialHooks549220{
					pick: func(*Object) *Object {
						if found {
							return target
						}
						return nil
					},
					trace: func(_, _ types.Pointf, flags MapTraceFlags) bool {
						if flags != 5 {
							t.Fatal("trace flags")
						}
						traces++
						return false
					},
				})
				want := 1
				if found || kind == MonsterStrikeWasp549220 {
					want = 0
				}
				wantTraces := 0
				if found {
					wantTraces = 1
				}
				if result != want || traces != wantTraces {
					t.Fatalf("miss=%d traces=%d", result, traces)
				}
			})
		}
	}
}

func TestMonsterStrikeSpecial549220CachedDamageLivePoisonOrder(t *testing.T) {
	for _, tc := range []struct {
		kind        MonsterStrikeKind549220
		message     strman.ID
		poisonFirst bool
	}{
		{MonsterStrikeScorpion549220, "aifunc.c:PoisonedByScorpion", false},
		{MonsterStrikeZombie549220, "aifunc.c:PoisonedByZombie", false},
		{MonsterStrikeWasp549220, "aifunc.c:PoisonedByWasp", true},
	} {
		t.Run(fmt.Sprint(tc.kind), func(t *testing.T) {
			unit := meleeMonsterTestObject532130(t)
			cached := unit.UpdateDataMonster()
			poisonDef := &MonsterDef{MeleeAttackPoisonChange136: 25}
			live := &MonsterUpdateData{MonsterDef: poisonDef}
			target := &Object{ObjClass: object.ClassPlayer}
			var events []string
			h := monsterStrikeSpecialHooks549220{
				MonsterStrikeSpecialRuntime549220: MonsterStrikeSpecialRuntime549220{
					MonsterStrikeDefaultRuntime549380: MonsterStrikeDefaultRuntime549380{
						Damage: func(v, source, weapon *Object, damage int, typ object.DamageType) bool {
							if v != target || source != unit || weapon != unit || damage != 6 || typ != 7 {
								t.Fatal("cached update damage")
							}
							cached.MonsterDef.MeleeAttackImpact120 = 9
							events = append(events, "damage")
							return true
						},
						ApplyForce: func(v *Object, origin types.Pointf, force float64) {
							if v != target || origin != unit.PosVec || force != 9 {
								t.Fatal("post-damage force")
							}
							events = append(events, "force")
						},
					},
					ActivatePoison: func(v *Object, strength, maximum int32) int32 {
						if v != target || strength != -2 || maximum != 7 {
							t.Fatal("cached poison pointer/live fields")
						}
						events = append(events, "poison")
						return 1
					},
					PriorityMessage: func(v *Object, message strman.ID, value byte) {
						if v != target || message != tc.message || value != 0 {
							t.Fatal("poison message")
						}
						events = append(events, "message")
					},
				},
				pick:  func(*Object) *Object { unit.UpdateData = unsafe.Pointer(live); return target },
				trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
				random: func(lo, hi int) int {
					if lo != 1 || hi != 100 {
						t.Fatal("poison RNG bounds")
					}
					live.MonsterDef = &MonsterDef{MeleeAttackPoisonChange136: 0}
					poisonDef.MeleeAttackPoisonChange136, poisonDef.MeleeAttackPoisonStrength140, poisonDef.MeleeAttackPoisonMax144 = 100, ^uint32(1), 7
					events = append(events, "random")
					return 80
				},
			}
			if monsterStrikeSpecial549220(unit, tc.kind, h) != 1 {
				t.Fatal("hit result")
			}
			want := []string{"damage", "force", "random", "poison", "message"}
			if tc.poisonFirst {
				want = []string{"damage", "random", "poison", "message", "force"}
			}
			assertMeleeEvents549380(t, events, want)
		})
	}
}

func TestMonsterStrikeSpecial549220AreaLiveClassForceAndQuake(t *testing.T) {
	for _, kind := range []MonsterStrikeKind549220{MonsterStrikeOgre549220, MonsterStrikeStoneGolem549220, MonsterStrikeMechGolem549220} {
		for _, impact := range []float32{-2, 0, 2, float32(math.NaN())} {
			t.Run(fmt.Sprintf("kind-%d/impact-%g", kind, impact), func(t *testing.T) {
				unit := meleeMonsterTestObject532130(t)
				unit.PosVec = types.Ptf(100, 200)
				target := &Object{PosVec: types.Ptf(110, 200)}
				update := unit.UpdateDataMonster()
				var events []string
				var result uint32
				golem := kind != MonsterStrikeOgre549220
				h := monsterStrikeSpecialHooks549220{
					MonsterStrikeSpecialRuntime549220: MonsterStrikeSpecialRuntime549220{
						MonsterStrikeDefaultRuntime549380: MonsterStrikeDefaultRuntime549380{
							Damage: func(v, source, weapon *Object, damage int, typ object.DamageType) bool {
								if v != target || source != unit || weapon != unit || damage != 6 || typ != 7 {
									t.Fatal("area damage identity")
								}
								target.ObjClass = object.ClassMonster
								update.MonsterDef.MeleeAttackImpact120 = impact
								events = append(events, "damage")
								return true
							},
							ApplyForce: func(v *Object, origin types.Pointf, force float64) {
								if v != target || origin != unit.PosVec || math.Float32bits(float32(force)) != math.Float32bits(impact) {
									t.Fatal("area live impact")
								}
								events = append(events, "force")
							},
						},
						Direction: func(Dir16) types.Pointf { return types.Ptf(1, 0) },
						Facing:    func(types.Pointf, int16, types.Pointf) int { return 1 },
						Distance:  func(*Object, *Object) float32 { return 15 }, // inclusive original boundary
						SetAreaResult: func(g bool, value uint32) {
							if g != golem {
								t.Fatal("area result bank")
							}
							result = value
							events = append(events, fmt.Sprintf("result-%d", value))
						},
						AreaResult: func(g bool) uint32 {
							if g != golem {
								t.Fatal("area result read")
							}
							return result
						},
						SetGolemMode: func(mode uint32) {
							want := uint32(0)
							if kind == MonsterStrikeMechGolem549220 {
								want = 1
							}
							if mode != want {
								t.Fatal("golem mode")
							}
							events = append(events, "mode")
						},
					},
					each: func(origin types.Pointf, radius float32, fn func(*Object) bool) {
						if origin != unit.PosVec || radius != 45 {
							t.Fatal("area radius")
						}
						fn(unit)
						fn(target)
					},
					trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool { events = append(events, "trace"); return true },
					quake: func(origin types.Pointf, strength int) {
						if origin != unit.PosVec || strength != 30 {
							t.Fatal("quake")
						}
						events = append(events, "quake")
					},
				}
				if monsterStrikeSpecial549220(unit, kind, h) != 1 {
					t.Fatal("area live class result")
				}
				var want []string
				if golem {
					want = append(want, "mode")
				}
				want = append(want, "result-0", "trace", "damage")
				if golem {
					want = append(want, "result-1")
				}
				if !golem || impact > 0 {
					want = append(want, "force")
				}
				if !golem {
					want = append(want, "result-1")
				} else {
					want = append(want, "quake")
				}
				assertMeleeEvents549380(t, events, want)
			})
		}
	}
}
