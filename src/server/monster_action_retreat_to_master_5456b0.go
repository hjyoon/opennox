package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionRetreatToMasterHooks5456D0 struct {
	push func(ai.ActionType, ...any) *AIStackItem
	pop  func() int
}

// monsterRetreatToMasterActive545580 restores GAME.EXE 00545580: retreating
// continues while health is below ResumeLevel, or while a spell-casting
// monster is prevented from casting by Anti-Magic.
func monsterRetreatToMasterActive545580(unit *Object) bool {
	if !monsterCanResumeAttack545520(unit) {
		return true
	}
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	return update.StatusFlags.Has(object.MonStatusCanCastSpells) && unit.HasEnchant(ENCHANT_ANTI_MAGIC)
}

// monsterActionRetreatToMaster5456D0 restores GAME.EXE 005456D0 with
// native-width owner and AI-stack pointers. When the owner is farther than the
// configured sight range plus 30 units, it schedules the original distance
// dependency followed by MOVE_TO.
func monsterActionRetreatToMaster5456D0(unit *Object, hooks monsterActionRetreatToMasterHooks5456D0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_RETREAT_TO_MASTER {
		return false
	}
	owner := unit.ObjOwner
	if owner == nil || !monsterRetreatToMasterActive545580(unit) {
		hooks.pop()
		return true
	}
	dx := float64(unit.PosVec.X - owner.PosVec.X)
	dy := float64(unit.PosVec.Y - owner.PosVec.Y)
	radius := float64(update.Field329) + 30
	if radius*radius < dx*dx+dy*dy {
		hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN, update.Field329, uint32(0), owner)
		hooks.push(ai.ACTION_MOVE_TO, owner.PosVec, owner)
	}
	return true
}

// MonsterActionRetreatToMaster5456D0 binds the restored update to the live
// native-width action stack.
func (s *Server) MonsterActionRetreatToMaster5456D0(unit *Object) bool {
	return monsterActionRetreatToMaster5456D0(unit, monsterActionRetreatToMasterHooks5456D0{
		push: unit.MonsterPushAction,
		pop:  unit.MonsterPopAction,
	})
}
