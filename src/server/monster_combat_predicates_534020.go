package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func MonsterFlying534020(unit *Object) bool {
	return unit != nil && uint32(unit.ObjSubClass)&0x400 != 0
}

func MonsterFacingDirection534120(unit *Object, direction types.Pointf) bool {
	return unit != nil && monsterFacingDirection534120(unit, direction)
}

func MonsterCanMelee534220(unit *Object) bool {
	return unit != nil && unit.UpdateData != nil && monsterFightCanMelee534220(unit, unit.UpdateDataMonster())
}

func MonsterCanShoot534280(unit *Object) bool {
	return unit != nil && unit.UpdateData != nil && monsterFightCanShoot534280(unit, unit.UpdateDataMonster())
}

func MonsterHasShield5342C0(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if unit.SubClass().Has(object.SubClass(object.MonsterNPC)) {
		return update.ArmorEquipFlags&0x3000000 != 0
	}
	return uint32(update.StatusFlags)&4 != 0
}

func MonsterCanCast534300(unit *Object) bool {
	return unit != nil && unit.UpdateData != nil &&
		unit.UpdateDataMonster().StatusFlags.Has(object.MonStatusCanCastSpells)
}

func MonsterIsMoving534320(unit *Object) bool {
	return unit != nil && unit.SpeedCur >= math.Float32frombits(0x3c23d70a)
}

func MonsterHasInterruptibleAction534340(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	var action uint32
	if update.AIStackInd == -1 {
		// GAME.EXE's signed head -1 reads the DWORD immediately before
		// the stack prefix, PE32 field132. Keep that native field identity.
		action = update.Field132
	} else if update.AIStackInd >= 0 && int(update.AIStackInd) < len(update.AIStack) {
		action = update.AIStack[update.AIStackInd].Action
	} else {
		return false
	}
	switch action {
	case 0, 1, 4, 23, 25, 26, 27:
		return true
	default:
		return false
	}
}

func MonsterLowAggression534400(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	value := unit.UpdateDataMonster().Aggression
	return value < math.Float32frombits(0x3ea8f5c3) && value > math.Float32frombits(0x3da3d70a)
}

func MonsterMeleeRange534470(unit *Object) float64 {
	if unit == nil || unit.UpdateData == nil || unit.UpdateDataMonster().MonsterDef == nil {
		return 0
	}
	return float64(unit.UpdateDataMonster().MonsterDef.MeleeAttackRange112)
}
