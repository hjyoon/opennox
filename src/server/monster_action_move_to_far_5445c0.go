package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterActionFarMoveTo5445C0 follows the original 005445C0 wrapper. In
// particular, a FIGHT push precedes the shared MOVE_TO body, so that body
// reads the new stack head rather than the old FAR_MOVE_TO action.
func monsterActionFarMoveTo5445C0(unit *Object, hooks monsterActionMoveToHooks5443F0, noticeThreat func() int) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		noticeThreat == nil || hooks.frame == nil || hooks.tickRate == nil ||
		hooks.random == nil || hooks.setMovePath == nil || hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if head := update.AIStackHead(); head == nil || head.Type() != ai.ACTION_FAR_MOVE_TO {
		return false
	}

	if (unit.Sub_5343C0() || unit.Nox_xxx_monsterCanAttackAtWill_534390()) && noticeThreat() != 0 {
		return true
	}
	if unit.Nox_xxx_monsterCanAttackAtWill_534390() && update.CurrentEnemy != nil {
		if fight := hooks.push(ai.ACTION_FIGHT); fight != nil {
			// The PE32 body reloads CurrentEnemy after the push. A callback may
			// clear it; discard the incomplete action instead of dereferencing nil.
			if enemy := update.CurrentEnemy; enemy != nil {
				fight.SetArgs(enemy.PosVec, hooks.frame())
			} else {
				hooks.pop()
			}
		}
	}
	// The original wrapper unconditionally invokes 005443F0. It observes
	// whichever action is now on top, including a newly pushed FIGHT.
	head := update.AIStackHead()
	if head == nil {
		return true
	}
	return monsterActionMoveToForAction5443F0(unit, head.Type(), hooks)
}

// MonsterActionFarMoveTo5445C0 binds the original wrapper to native-width
// Object and MonsterUpdateData pointers instead of its PE32 integer fields.
func (s *Server) MonsterActionFarMoveTo5445C0(unit *Object, setDetailedPath func(*Object, *types.Pointf)) bool {
	if unit == nil {
		return false
	}
	return monsterActionFarMoveTo5445C0(unit, s.monsterActionMoveToHooks5443F0(unit, setDetailedPath), unit.Sub_545E60)
}
