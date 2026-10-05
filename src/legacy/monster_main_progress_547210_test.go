package legacy

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Use C-owned native records, the production MainAI wrapper, server logic
// RNG, and actual wall/obstacle/tile services. This proves bookkeeping and
// stack scheduling, not collision-driven world movement or bot gameplay.
func TestMonsterMainProgress547210ActualLegacyBinding(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE} {
		for _, mode := range []string{"moved", "before-timeout", "stalled-in-fight"} {
			t.Run(fmt.Sprintf("%s/%s", action, mode), func(t *testing.T) {
				srv, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
				unit.SpeedCur = 10
				update.AIStackInd = 1
				update.AIStack[1].Action = uint32(action)
				update.Field124 = 1385
				update.Field125, update.Field126 = math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)
				switch mode {
				case "moved":
					update.Field125, update.Field126 = math.Float32bits(100), math.Float32bits(100)
				case "stalled-in-fight":
					update.Field124 = 1384
				}
				before := *update
				t.Logf("actual common MainAI: unit=%p update=%p action=%v mode=%s", unit, update, action, mode)
				Nox_xxx_monsterMainAIFn_547210(unit)
				if unit.HealthData.Cur != 100 {
					t.Fatal("movement tail changed health")
				}
				if mode == "before-timeout" {
					if *update != before || srv.AI.StackChanged {
						t.Fatal("half-FPS equality scheduled frustration or changed bookkeeping")
					}
					return
				}
				if update.Field124 != 1400 || update.Field125 != math.Float32bits(300) || update.Field126 != math.Float32bits(300) {
					t.Fatal("native movement record was not updated")
				}
				if mode == "moved" {
					before.Field124, before.Field125, before.Field126 = update.Field124, update.Field125, update.Field126
					if *update != before || srv.AI.StackChanged {
						t.Fatal("displacement changed fields outside the three progress DWORDs")
					}
					return
				}
				index := int(update.AIStackInd)
				head := update.AIStackHead()
				position := head.ArgPos(0)
				if !update.StatusFlags.Has(object.MonStatusFrustrated) || !srv.AI.StackChanged || index < 2 ||
					update.AIStack[0].Type() != ai.ACTION_FIGHT || update.AIStack[index-1].Type() != ai.DEPENDENCY_TIME ||
					update.AIStack[index-1].ArgU32(0) != 1431 || head.Type() != ai.ACTION_DODGE ||
					position.X != 300 || math.Abs(float64(position.Y-300)) != 15 || head.ArgU32(2) != 0 {
					t.Fatalf("real frustration services did not schedule lateral dodge: stack=%+v", update.GetAIStack())
				}
				if action == ai.ACTION_FLEE && update.Field127 != 1400 {
					t.Fatal("live FLEE stack did not update move-attempt frame")
				}
			})
		}
	}
}
