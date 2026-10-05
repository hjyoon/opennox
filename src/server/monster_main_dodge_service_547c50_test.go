package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterMainDodgeServiceFixture547C50(t *testing.T) (*Server, *Object, *MonsterUpdateData, MonsterMainRuntime547210) {
	t.Helper()
	s, unit, update := monsterMainHealthRetreatFixture547210(t)
	unit.PosVec, unit.NewPos = types.Ptf(100, 200), types.Ptf(100, 200)
	unit.Direction1, unit.SpeedCur = 0, 10
	unit.HealthData.Cur, unit.HealthData.Max = 100, 100
	update.RetreatLevel = 0
	return s, unit, update, MonsterMainRuntime547210{
		RandomFloat: func(min, max float32) float64 {
			if min != 2 || max != 3 {
				t.Fatalf("float bounds=%g..%g", min, max)
			}
			return 2
		},
		RandomInt: func(min, max int) int {
			if min != 0 || max != 100 {
				t.Fatalf("integer bounds=%d..%d", min, max)
			}
			return 50
		},
		TraceRay: func(_, _ types.Pointf, flags MapTraceFlags) bool {
			if flags != MapTraceFlag1 {
				t.Fatalf("ray flags=%#x", flags)
			}
			return true
		},
		TraceObstacles: func(got *Object, _, _ types.Pointf) bool {
			if got != unit {
				t.Fatalf("obstacle unit=%p, want %p", got, unit)
			}
			return true
		},
		TileAt: func(types.Pointf) int { return 0 },
	}
}

func TestMonsterMainDodgeService547C50CachedOriginAcrossRNG(t *testing.T) {
	for _, stage := range []string{"float", "integer"} {
		t.Run(stage, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			origin, destination := unit.PosVec, types.Ptf(300, 415)
			move := func() { unit.PosVec, unit.Direction1 = types.Ptf(300, 400), 64 }
			if stage == "float" {
				runtime.RandomFloat = func(_, _ float32) float64 { move(); return 2 }
			} else {
				runtime.RandomInt = func(_, _ int) int { move(); return 50 }
			}
			runtime.TraceRay = func(from, to types.Pointf, _ MapTraceFlags) bool {
				if from != origin || to != destination {
					t.Fatalf("ray=%v->%v, want cached origin %v->live destination %v", from, to, origin, destination)
				}
				return true
			}
			runtime.TraceObstacles = func(_ *Object, from, to types.Pointf) bool {
				if from != origin || to != destination {
					t.Fatalf("obstacles=%v->%v, want %v->%v", from, to, origin, destination)
				}
				return true
			}
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) || update.AIStackHead().ArgPos(0) != destination {
				t.Fatal("dodge did not preserve entry direction and post-RNG destination")
			}
		})
	}
}

func TestMonsterMainDodgeService547C50CachedOriginAcrossFailedProbe(t *testing.T) {
	for _, stage := range []string{"ray", "obstacles", "lava"} {
		t.Run(stage, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			origin, attempt := unit.PosVec, 0
			move := func() { unit.PosVec, unit.Direction1 = types.Ptf(300, 400), 64 }
			wantTo := func() types.Pointf {
				if attempt == 1 {
					return types.Ptf(100, 215)
				}
				return types.Ptf(300, 415)
			}
			runtime.RandomFloat = func(_, _ float32) float64 { attempt++; return 2 }
			runtime.TraceRay = func(from, to types.Pointf, _ MapTraceFlags) bool {
				if from != origin || to != wantTo() {
					t.Fatalf("attempt %d ray=%v->%v, want %v->%v", attempt, from, to, origin, wantTo())
				}
				if attempt == 1 && stage == "ray" {
					move()
					return false
				}
				return true
			}
			runtime.TraceObstacles = func(_ *Object, from, to types.Pointf) bool {
				if from != origin || to != wantTo() {
					t.Fatalf("attempt %d obstacle origin changed: %v->%v", attempt, from, to)
				}
				if attempt == 1 && stage == "obstacles" {
					move()
					return false
				}
				return true
			}
			runtime.TileAt = func(to types.Pointf) int {
				if to != wantTo() {
					t.Fatalf("attempt %d tile point=%v, want %v", attempt, to, wantTo())
				}
				if attempt == 1 && stage == "lava" {
					move()
					return 6
				}
				return 0
			}
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) || attempt != 2 || update.AIStackHead().ArgPos(0) != types.Ptf(300, 415) {
				t.Fatal("second attempt did not retain original ray start and direction")
			}
		})
	}
}

func TestMonsterMainDodgeService547C50CachedOriginAcrossSuccessfulTrace(t *testing.T) {
	for _, stage := range []string{"ray", "obstacles"} {
		t.Run(stage, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			origin, destination := unit.PosVec, types.Ptf(100, 215)
			move := func() { unit.PosVec = types.Ptf(300, 400) }
			runtime.TraceRay = func(from, to types.Pointf, _ MapTraceFlags) bool {
				if from != origin || to != destination {
					t.Fatal("ray did not use cached line")
				}
				if stage == "ray" {
					move()
				}
				return true
			}
			runtime.TraceObstacles = func(_ *Object, from, to types.Pointf) bool {
				if from != origin || to != destination {
					t.Fatalf("successful ray changed obstacle line: %v->%v", from, to)
				}
				if stage == "obstacles" {
					move()
				}
				return true
			}
			runtime.TileAt = func(to types.Pointf) int {
				if to != destination {
					t.Fatal("tile point changed after obstacle callback")
				}
				return 0
			}
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) || update.AIStackHead().ArgPos(0) != destination {
				t.Fatal("scheduled destination changed")
			}
		})
	}
}

func TestMonsterMainDodgeService547C50PostPushDeadline(t *testing.T) {
	for _, wrap := range []bool{false, true} {
		t.Run(fmt.Sprint(wrap), func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			update.AIStack[0].Field5 = 1
			old, calls := aiActions[ai.ACTION_FIGHT], 0
			aiActions[ai.ACTION_FIGHT] = monsterMainFearCancel547210{cancel: func(*Object) {
				calls++
				s.SetFrame(201)
				s.SetTickRate(50)
				if wrap {
					s.SetFrame(^uint32(0) - 2)
					s.SetTickRate(9)
				}
				unit.PosVec = types.Ptf(300, 400)
			}}
			t.Cleanup(func() {
				if old == nil {
					delete(aiActions, ai.ACTION_FIGHT)
				} else {
					aiActions[ai.ACTION_FIGHT] = old
				}
			})
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) {
				t.Fatal("clear dodge rejected")
			}
			deadline := uint32(251)
			if wrap {
				deadline = 6
			}
			if calls != 1 || update.AIStackInd != 2 || update.AIStack[1].Type() != ai.DEPENDENCY_TIME ||
				update.AIStack[1].ArgU32(0) != deadline || update.AIStackHead().Type() != ai.ACTION_DODGE ||
				update.AIStackHead().ArgPos(0) != types.Ptf(100, 215) || update.AIStackHead().ArgU32(2) != 0 {
				t.Fatalf("deadline was not read after normal push: calls=%d stack=%+v, want %d", calls, update.GetAIStack(), deadline)
			}
		})
	}
}

func TestMonsterMainDodgeService547C50FiveAttemptsAndShortCircuit(t *testing.T) {
	for _, stage := range []string{"ray", "obstacles", "lava"} {
		for _, successAt := range []int{1, 5, 0} {
			t.Run(fmt.Sprintf("%s/%d", stage, successAt), func(t *testing.T) {
				s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
				attempt, events, want := 0, []string{}, []string{}
				runtime.RandomFloat = func(_, _ float32) float64 { attempt++; events = append(events, "float"); return 2 }
				runtime.RandomInt = func(_, _ int) int { events = append(events, "integer"); return 50 }
				runtime.TraceRay = func(_, _ types.Pointf, _ MapTraceFlags) bool {
					events = append(events, "ray")
					return stage != "ray" || attempt == successAt
				}
				runtime.TraceObstacles = func(_ *Object, _, _ types.Pointf) bool {
					events = append(events, "obstacles")
					return stage != "obstacles" || attempt == successAt
				}
				runtime.TileAt = func(types.Pointf) int {
					events = append(events, "tile")
					if stage == "lava" && attempt != successAt {
						return 6
					}
					return 0
				}
				tries := successAt
				if tries == 0 {
					tries = 5
				}
				for i := 1; i <= tries; i++ {
					want = append(want, "float", "integer", "ray")
					if stage == "ray" && i != successAt {
						continue
					}
					want = append(want, "obstacles")
					if stage == "obstacles" && i != successAt {
						continue
					}
					want = append(want, "tile")
				}
				if got := s.monsterMainCheckDodgeables547C50(unit, runtime); got != (successAt != 0) || attempt != tries || !reflect.DeepEqual(events, want) {
					t.Fatalf("got=%t attempts=%d events=%v, want attempts=%d %v", got, attempt, events, tries, want)
				}
				if successAt == 0 && (update.AIStackInd != 0 || update.AIStackHead().Type() != ai.ACTION_FIGHT) {
					t.Fatal("five blocked destinations changed action stack")
				}
			})
		}
	}
}

func TestMonsterMainDodgeService547C50FloatSpillCapAndDirection(t *testing.T) {
	for _, tt := range []struct {
		name   string
		random float64
		speed  float32
		side   int
		wantY  float32
	}{
		{"right", 2, 4, 50, 8}, {"left", 2, 4, 49, -8}, {"cap", 3, 10, 100, 15},
		{"cap-left", 3, 10, 0, -15}, {"negative-speed-not-clamped", 2, -10, 50, -20},
		{"binary32-spill", 2.00000005, 4, 50, 8}, {"unordered", math.NaN(), 4, 50, float32(math.NaN())},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			unit.PosVec, unit.SpeedCur = types.Ptf(0, 0), tt.speed
			runtime.RandomFloat = func(_, _ float32) float64 { return tt.random }
			runtime.RandomInt = func(_, _ int) int { return tt.side }
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) {
				t.Fatal("probe rejected")
			}
			got := update.AIStackHead().ArgPos(0).Y
			if math.IsNaN(float64(tt.wantY)) {
				if !math.IsNaN(float64(got)) {
					t.Fatal("unordered spill lost")
				}
			} else if math.Float32bits(got) != math.Float32bits(tt.wantY) {
				t.Fatalf("destination Y=%g bits=%#x, want %g", got, math.Float32bits(got), tt.wantY)
			}
		})
	}
}

func TestMonsterMainDodgeService547C50RejectedPushStillSucceeds(t *testing.T) {
	for _, mode := range []string{"full-stack", "dead-head"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			if mode == "full-stack" {
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStackHead().Action = uint32(ai.ACTION_WAIT)
			} else {
				update.AIStackHead().Action = uint32(ai.ACTION_DEAD)
			}
			before := *update
			if !s.monsterMainCheckDodgeables547C50(unit, runtime) || *update != before {
				t.Fatal("original helper must return success even when both normal pushes reject")
			}
		})
	}
}
