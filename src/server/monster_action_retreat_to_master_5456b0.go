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
	// 005456FB reloads the owner after the retreat predicate. FSUB/FADD
	// and Y-square/X-square/sum retain precision 53 / ToZero until
	// FCOMPP, with no binary32 spill or contracted distance sum.
	owner = unit.ObjOwner
	dx := monsterMoveToRunAddChop53_544434(float64(unit.PosVec.X), -float64(owner.PosVec.X))
	dy := monsterMoveToRunAddChop53_544434(float64(unit.PosVec.Y), -float64(owner.PosVec.Y))
	radius := monsterMoveToRunAddChop53_544434(float64(update.Field329), 30)
	ySquared := monsterMoveToRunSquareChop53_544434(dy)
	xSquared := monsterMoveToRunSquareChop53_544434(dx)
	distance := monsterMoveToRunAddChop53_544434(ySquared, xSquared)
	// 0054572F tests C0 alone: less and unordered both schedule movement.
	if !(monsterMoveToRunSquareChop53_544434(radius) >= distance) {
		if dependency := hooks.push(ai.DEPENDENCY_OBJECT_FARTHER_THAN); dependency != nil {
			// Push may invoke Cancel. The range remains entry-cached, but
			// the owner is live; 00545745..00545754 writes no Args[1].
			dependency.SetArgs(update.Field329)
			dependency.Args[2] = uintptr(unit.ObjOwner.CObj())
		}
		if move := hooks.push(ai.ACTION_MOVE_TO); move != nil {
			// 00545766..0054577F loads raw X/Y after successful push,
			// then reloads the native owner identity for the final store.
			move.SetArgs(unit.ObjOwner.PosVec)
			move.Args[2] = uintptr(unit.ObjOwner.CObj())
		}
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
