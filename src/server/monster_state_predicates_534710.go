package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func MonsterHoldPosition534710(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.StatusFlags.Has(object.MonStatusHoldYourGround) {
		return true
	}
	if update.AIStackInd == -1 {
		// The signed original index reads PE32 field132 before the stack.
		return update.Field132 == uint32(ai.ACTION_GUARD)
	}
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	return update.AIStack[update.AIStackInd].Action == uint32(ai.ACTION_GUARD)
}

func MonsterStatusBit95347A0(unit *Object) bool {
	return unit != nil && unit.UpdateData != nil && uint32((*MonsterUpdateData)(unit.UpdateData).StatusFlags)&0x200 != 0
}

func UnitInjured5347C0(unit *Object) bool {
	return unit != nil && unit.HealthData != nil && unit.HealthData.Max != 0 && unit.HealthData.Cur < unit.HealthData.Max
}

func UnitPoisoned5347F0(unit *Object) bool {
	// The old public name says isNotPoisoned; GAME.EXE tests nonzero.
	return unit != nil && unit.Poison540 != 0
}
