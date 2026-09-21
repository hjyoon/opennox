package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterActionGet50A020 returns the current monster action through the
// native-width AI stack. Valid GAME.EXE callers always have a live stack;
// malformed or absent native state is rejected deterministically here.
func (obj *Object) MonsterActionGet50A020() ai.ActionType {
	if obj == nil || obj.UpdateData == nil {
		return ai.ACTION_INVALID
	}
	ud := obj.UpdateDataMonster()
	ind := int(ud.AIStackInd)
	if ind < 0 || ind >= len(ud.AIStack) {
		return ai.ACTION_INVALID
	}
	return ud.AIStack[ind].Type()
}

// MonsterActionPrevious50A040 walks below the current action and skips every
// condition according to GAME.EXE 0050A010's signed action-ID comparison.
func (ud *MonsterUpdateData) MonsterActionPrevious50A040() ai.ActionType {
	if ud == nil {
		return ai.ACTION_INVALID
	}
	ind := int(ud.AIStackInd)
	if ind < 0 || ind >= len(ud.AIStack) {
		return ai.ACTION_INVALID
	}
	for i := ind - 1; i >= 0; i-- {
		if action := ud.AIStack[i].Type(); !action.IsCondition() {
			return action
		}
	}
	return ai.ACTION_INVALID
}

// MonsterActionPrevious50A040 is the Object boundary used by the legacy C
// symbol at 0050A040.
func (obj *Object) MonsterActionPrevious50A040() ai.ActionType {
	if obj == nil || obj.UpdateData == nil {
		return ai.ACTION_INVALID
	}
	return obj.UpdateDataMonster().MonsterActionPrevious50A040()
}

// MonsterActionPushIfChanged50A360 preserves the original monster-class gate
// and suppresses a push only when the current action is exactly equal.
func (obj *Object) MonsterActionPushIfChanged50A360(action ai.ActionType) *AIStackItem {
	if obj == nil || obj.UpdateData == nil || !obj.Class().Has(object.ClassMonster) {
		return nil
	}
	if obj.MonsterActionGet50A020() == action {
		return nil
	}
	return obj.MonsterPushAction(action)
}
