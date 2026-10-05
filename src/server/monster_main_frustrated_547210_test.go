package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// These cases exercise the call order in GAME.EXE 00547A85..00547B69,
// including the distinction between the entry-cached record and live stack.
func TestMonsterMainFrustrated547210PostPushClock(t *testing.T) {
	for _, mode := range []string{"normal", "frame-wrap", "signed-RNG-upper-bound"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO), Field5: 1}
			events := []string{}
			fps, frame, duration := uint32(50), uint32(501), 47
			if mode == "frame-wrap" {
				fps, frame, duration = 9, ^uint32(0)-5, 17
			} else if mode == "signed-RNG-upper-bound" {
				fps, duration = 0x80000001, 1
			}
			old := aiActions[ai.ACTION_MOVE_TO]
			aiActions[ai.ACTION_MOVE_TO] = monsterMainFearCancel547210{cancel: func(*Object) {
				events = append(events, "cancel")
				s.SetFrame(201)
				s.SetTickRate(fps)
				unit.PosVec = types.Ptf(300, 400)
			}}
			t.Cleanup(func() {
				if old == nil {
					delete(aiActions, ai.ACTION_MOVE_TO)
				} else {
					aiActions[ai.ACTION_MOVE_TO] = old
				}
			})
			runtime.RandomFloat = func(float32, float32) float64 { t.Fatal("WAIT admission reached dodge"); return 0 }
			runtime.RandomInt = func(min, max int) int {
				if len(events) == 0 {
					if min != 0 || max != 100 {
						t.Fatalf("admission bounds %d..%d", min, max)
					}
					events = append(events, "admit")
					return 33
				}
				if !reflect.DeepEqual(events, []string{"admit", "cancel"}) ||
					update.AIStackHead().Type() != ai.ACTION_WAIT || min != int(fps>>1) || max != int(int32(2*fps)) {
					t.Fatalf("duration must follow normal push: events=%v, head=%v, bounds=%d..%d", events, update.AIStackHead().Type(), min, max)
				}
				events = append(events, "duration")
				s.SetFrame(frame)
				s.SetTickRate(7)
				unit.PosVec = types.Ptf(-0.0, -123)
				unit.PosVec.X = math.Float32frombits(0x80000000)
				return duration
			}
			if !s.monsterMainFrustrated547210(unit, update, runtime) ||
				!reflect.DeepEqual(events, []string{"admit", "cancel", "duration"}) ||
				update.AIStackHead().ArgU32(0) != frame+uint32(duration) || update.Field124 != frame ||
				update.Field125 != 0x80000000 || update.Field126 != math.Float32bits(-123) ||
				!update.StatusFlags.Has(object.MonStatusFrustrated) {
				t.Fatalf("post-push clock/bookkeeping lost: events=%v update=%+v", events, update)
			}
		})
	}
}

func TestMonsterMainFrustrated547210RejectedPushSkipsDurationRNG(t *testing.T) {
	for _, mode := range []string{"full-stack", "dead-head", "class-changed-by-RNG"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
			update.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
			switch mode {
			case "full-stack":
				update.AIStackInd = int8(len(update.AIStack) - 1)
				update.AIStack[update.AIStackInd].Action = uint32(ai.ACTION_MOVE_TO)
			case "dead-head":
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			}
			stack, calls := update.AIStack, 0
			runtime.RandomInt = func(min, max int) int {
				calls++
				if calls != 1 || min != 0 || max != 100 {
					t.Fatalf("failed push consumed duration RNG: call %d, bounds %d..%d", calls, min, max)
				}
				if mode == "class-changed-by-RNG" {
					// Admission has already queried a genuine monster's live
					// stack. A later callback can make the normal push reject.
					unit.ObjClass = object.ClassPlayer
				}
				return 33
			}
			if !s.monsterMainFrustrated547210(unit, update, runtime) || calls != 1 || update.AIStack != stack ||
				update.Field124 != s.Frame() || update.Field125 != math.Float32bits(unit.PosVec.X) {
				t.Fatal("original failed-push return/bookkeeping changed")
			}
		})
	}
}

func TestMonsterMainFrustrated547210LiveRetreatAndFightQueries(t *testing.T) {
	for _, cachedAction := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_RETREAT, ai.ACTION_FIGHT} {
		for _, liveAction := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_RETREAT, ai.ACTION_RETREAT_TO_MASTER, ai.ACTION_FLEE, ai.ACTION_FIGHT} {
			t.Run(fmt.Sprintf("cached-%s/live-%s", cachedAction, liveAction), func(t *testing.T) {
				s, unit, cached, runtime := monsterMainDodgeServiceFixture547C50(t)
				cached.AIStack[0].Action = uint32(cachedAction)
				live := &MonsterUpdateData{AIStackInd: 1, Field127: 99}
				live.AIStack[0].Action = uint32(liveAction)
				live.AIStack[1].Action = uint32(ai.ACTION_MOVE_TO)
				unit.UpdateData = unsafe.Pointer(live)
				integers, floats := 0, 0
				runtime.RandomFloat = func(float32, float32) float64 { floats++; return 2 }
				runtime.TraceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false }
				runtime.RandomInt = func(min, max int) int {
					integers++
					if liveAction == ai.ACTION_FIGHT {
						if min != 0 || max != 100 {
							t.Fatalf("FIGHT used WAIT RNG %d..%d", min, max)
						}
						return 50
					}
					if integers == 1 {
						if min != 0 || max != 100 {
							t.Fatal("missing WAIT admission RNG")
						}
						return 33
					}
					if min != 15 || max != 60 {
						t.Fatalf("duration bounds %d..%d", min, max)
					}
					return 47
				}
				if !s.monsterMainFrustrated547210(unit, cached, runtime) {
					t.Fatal("frustration rejected")
				}
				if liveAction == ai.ACTION_FIGHT {
					if integers != 5 || floats != 5 || live.AIStackInd != 1 {
						t.Fatalf("FIGHT must try five dodges without WAIT: ints=%d floats=%d stack=%+v", integers, floats, live.GetAIStack())
					}
				} else if integers != 2 || floats != 0 || live.AIStackInd != 2 || live.AIStackHead().Type() != ai.ACTION_WAIT {
					t.Fatalf("live non-FIGHT branch lost WAIT: ints=%d floats=%d stack=%+v", integers, floats, live.GetAIStack())
				}
				wantRetreatFrame := uint32(0)
				if liveAction == ai.ACTION_RETREAT || liveAction == ai.ACTION_RETREAT_TO_MASTER || liveAction == ai.ACTION_FLEE {
					wantRetreatFrame = s.Frame()
				}
				if cached.Field127 != wantRetreatFrame || live.Field127 != 99 ||
					!cached.StatusFlags.Has(object.MonStatusFrustrated) || live.StatusFlags.Has(object.MonStatusFrustrated) ||
					cached.AIStackInd != 0 || cached.AIStack[0].Type() != cachedAction {
					t.Fatal("queries must use live stack, bookkeeping must use entry-cached record")
				}
			})
		}
	}
}

func TestMonsterMainFrustrated547210AdmissionBoundaryAndDodge(t *testing.T) {
	for _, admission := range []int{32, 33} {
		for _, clear := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/clear-%t", admission, clear), func(t *testing.T) {
				s, unit, update, runtime := monsterMainDodgeServiceFixture547C50(t)
				update.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
				floats, durationCalls, admitted := 0, 0, false
				runtime.RandomFloat = func(float32, float32) float64 { floats++; return 2 }
				runtime.TraceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { return clear }
				runtime.RandomInt = func(min, max int) int {
					if !admitted {
						admitted = true
						return admission
					}
					if min == 15 && max == 60 {
						durationCalls++
						return 47
					}
					if min != 0 || max != 100 {
						t.Fatalf("unexpected RNG bounds %d..%d", min, max)
					}
					return 50
				}
				if !s.monsterMainFrustrated547210(unit, update, runtime) {
					t.Fatal("frustration rejected")
				}
				wantFloats, wantDuration, wantHead := 0, 1, ai.ACTION_WAIT
				if admission == 32 {
					wantFloats = 5
					if clear {
						wantFloats, wantDuration, wantHead = 1, 0, ai.ACTION_DODGE
					}
				}
				if floats != wantFloats || durationCalls != wantDuration || update.AIStackHead().Type() != wantHead {
					t.Fatalf("admission/short circuit = floats %d, duration %d, head %v", floats, durationCalls, update.AIStackHead().Type())
				}
			})
		}
	}
}
