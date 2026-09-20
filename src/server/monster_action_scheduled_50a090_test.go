package server

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterActionScheduled50A090ExcludesCurrentAction(t *testing.T) {
	update := new(MonsterUpdateData)
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_FLEE)

	if update.HasScheduledAction(ai.ACTION_FLEE) {
		t.Fatal("current FLEE action was reported as scheduled")
	}
	if !update.HasAction(ai.ACTION_FLEE) {
		t.Fatal("HasAction no longer includes the current action")
	}
}

func TestMonsterActionScheduled50A090SearchesOnlyBelowHead(t *testing.T) {
	update := new(MonsterUpdateData)
	update.AIStackInd = 3
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
	update.AIStack[2].Action = uint32(ai.ACTION_FLEE)
	update.AIStack[3].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
	// This slot is outside the live stack and must not be observed.
	update.AIStack[4].Action = uint32(ai.ACTION_HUNT)

	for _, action := range []ai.ActionType{ai.ACTION_GUARD, ai.ACTION_FIGHT, ai.ACTION_FLEE} {
		if !update.HasScheduledAction(action) {
			t.Errorf("queued action %s was not reported as scheduled", action)
		}
	}
	for _, action := range []ai.ActionType{ai.ACTION_CAST_SPELL_ON_OBJECT, ai.ACTION_HUNT} {
		if update.HasScheduledAction(action) {
			t.Errorf("non-queued action %s was reported as scheduled", action)
		}
	}
}

func TestMonsterActionScheduled50A090EmptyStack(t *testing.T) {
	update := new(MonsterUpdateData)
	update.AIStackInd = -1
	if update.HasScheduledAction(ai.ACTION_FIGHT) {
		t.Fatal("empty action stack reported a scheduled action")
	}

	var nilUpdate *MonsterUpdateData
	if nilUpdate.HasScheduledAction(ai.ACTION_FIGHT) {
		t.Fatal("nil monster update data reported a scheduled action")
	}
}
