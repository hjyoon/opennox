package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestAIGuardLocation546010NativeDependencyPipeline(t *testing.T) {
	old := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(old) })
	s, unit, enemy, health := aiAliveDependencyNative546A70(t)
	guard := types.Ptf(1406, 1792)
	health.Cur, health.Max = 30, 30
	for _, tc := range []struct {
		name   string
		pos    types.Pointf
		recent bool
		keep   bool
	}{
		{"original-far-recognition-at-home", guard, false, true},
		{"inside-guard-radius", types.Ptf(2155, 1792), false, true},
		{"equal-guard-radius", types.Ptf(2156, 1792), false, true},
		{"outside-guard-radius", types.Ptf(2157, 1792), false, false},
		{"recent-hit-outside-radius", types.Ptf(2157, 1792), true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s.SetFrame(702) // Past the original ten-second UNDER_ATTACK window.
			unit.ObjFlags, unit.ObjClass = object.FlagActive|object.FlagEnabled, object.ClassMonster
			unit.TypeInd = 1 // Avoid the unloaded registry's zero-valued Plant/Mimic IDs.
			unit.PosVec, unit.PrevPos, unit.NetCode = tc.pos, tc.pos, 0
			enemy.PosVec = types.Ptf(1278, 2147)
			update := unit.UpdateDataMonster()
			*update = server.MonsterUpdateData{Aggression: 0.5, SightRange: 500, CurrentEnemy: enemy}
			update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_GUARD), Field5: 1}
			update.AIStack[0].SetArgs(guard, uint32(96))
			// The original guard writer must not overwrite this unused slot.
			update.AIStack[2].Args[1] = uintptr(unsafe.Pointer(enemy))
			s.S().MonsterActionGuard546010(unit, func(*server.Object) int { return 0 })
			if update.AIStackInd != 4 || update.AIStackHead().Type() != ai.ACTION_FIGHT {
				t.Fatalf("original guard did not schedule FIGHT: %+v", update.GetAIStack())
			}
			if tc.recent {
				update.AIStack[1].Args[0] = uintptr(s.Frame() - 1)
			}
			before := *update
			s.AI.StackChanged = false
			(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
			if tc.keep {
				if *update != before || s.AI.StackChanged {
					t.Fatalf("guard FIGHT was cancelled before it could start: index=%d stack=%+v", update.AIStackInd, update.GetAIStack())
				}
			} else if update.AIStackInd != 0 || update.AIStackHead().Type() != ai.ACTION_GUARD || !s.AI.StackChanged {
				t.Fatalf("original guard leash did not cancel FIGHT: %+v", update.GetAIStack())
			}
			if update.CurrentEnemy != enemy || health.Cur != 30 || health.Max != 30 {
				t.Fatal("guard/dependency pipeline changed the detected enemy or its health")
			}
		})
	}
}
