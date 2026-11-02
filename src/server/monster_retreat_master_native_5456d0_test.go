package server

import (
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Ordinary callers already treat missing health as "continue retreating";
// the original raw public entry faults instead. Sharing the update core must
// retain both admission contracts, rather than changing one to match the other.
func TestMonsterRetreatMaster5456D0OrdinaryGuardedHealthAndRawFault(t *testing.T) {
	_, unit, update, owners := monsterRetreatMasterNativeFixture5456D0(t)
	unit.PosVec, owners[0].PosVec = types.Ptf(0, 0), types.Ptf(200, 0)
	unit.HealthData, update.Field329 = nil, 20
	var actions []ai.ActionType
	pop := func() int { t.Fatal("guarded injured owner path unexpectedly popped"); return 0 }
	if !monsterActionRetreatToMaster5456D0(unit, monsterActionRetreatToMasterHooks5456D0{
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			if len(args) != 0 {
				t.Fatal("payload evaluated before Push")
			}
			actions = append(actions, action)
			return nil
		},
		pop: pop,
	}) || len(actions) != 2 || actions[0] != ai.DEPENDENCY_OBJECT_FARTHER_THAN || actions[1] != ai.ACTION_MOVE_TO {
		t.Fatal("existing ordinary guarded-health contract changed", actions)
	}
	var fault any
	func() {
		defer func() { fault = recover() }()
		unit.MonsterRetreatToMasterNative5456D0(func(ai.ActionType) *AIStackItem {
			t.Fatal("raw entry reached Push before original null-health fault")
			return nil
		}, pop)
	}()
	if fault == nil {
		t.Fatal("raw original null-health fault replaced by ordinary admission")
	}
}
