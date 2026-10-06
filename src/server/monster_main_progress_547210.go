package server

import (
	"math"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// These availability checks only contain incomplete runtimes. Unlike the
// old staged-port classifiers, they add no unit-state gates when services
// are bound, and never call a missile probe or consume RNG themselves.
func monsterMainBlockServicesAvailable547210(unit *Object, update *MonsterUpdateData, head *AIStackItem, runtime MonsterMainRuntime547210) bool {
	if runtime.TestShield != nil || head == nil || head.Type() == ai.ACTION_MELEE_ATTACK || head.Type() == ai.ACTION_MISSILE_ATTACK {
		return true
	}
	return !(unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) && update.WeaponEquipFlags&0x400 != 0) &&
		!monsterMainHasShield5342C0(unit)
}

func (s *Server) monsterMainDodgeServicesAvailable547210(unit *Object, update *MonsterUpdateData, runtime MonsterMainRuntime547210) bool {
	if !noxflags.HasGame(noxflags.GameModeCoop) ||
		!(unit.UpdateDataMonster().Aggression >= monsterMainPassiveAggressionLimit547210) ||
		unit.HasEnchant(ENCHANT_CONFUSED) || update.MonsterDef == nil ||
		!update.MonsterDef.StatusFlags92.Has(object.MonStatusCanDodge) ||
		unit.UpdateDataMonster().HasAction(ai.ACTION_DODGE) {
		return true
	}
	return runtime.TestShield != nil && runtime.TileAt != nil &&
		((runtime.RandomInt != nil && runtime.RandomFloat != nil) || s.Rand.Logic != nil)
}

// 005479FA..00547C45 is shared by every state reaching the tail, not just
// passive/active classifiers. The stack ITEM pointer and progress record are
// cached by MainAI before callbacks, but its action is read only here. x87
// retains the differences and squared sum instead of spilling to binary32.
// A stalled tick always returns before food and bot weapon searches.
func (s *Server) monsterMainProgressTail547210(unit *Object, update *MonsterUpdateData, head *AIStackItem, runtime MonsterMainRuntime547210) bool {
	switch head.Type() {
	case ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE:
		// 00547A19..00547A35 retains each operation at precision 53.
		// Gameplay's 0043E2C1 control word selects round-toward-zero;
		// preserve those boundaries without contracting the sum into FMA.
		dx := monsterMoveToRunAddChop53_544434(float64(math.Float32frombits(update.Field125)), -float64(unit.PosVec.X))
		dy := monsterMoveToRunAddChop53_544434(float64(math.Float32frombits(update.Field126)), -float64(unit.PosVec.Y))
		ySquared := monsterMoveToRunSquareChop53_544434(dy)
		xSquared := monsterMoveToRunSquareChop53_544434(dx)
		if monsterMoveToRunAddChop53_544434(ySquared, xSquared) > 225 {
			update.Field124 = s.Frame()
			update.Field125 = math.Float32bits(unit.PosVec.X)
			update.Field126 = math.Float32bits(unit.PosVec.Y)
		} else if s.Frame()-update.Field124 > s.TickRate()>>1 {
			return s.monsterMainFrustrated547210(unit, update, runtime)
		}
	}
	if !s.monsterMainEatNearbyFood547210(unit, update, runtime) {
		return false
	}
	return s.monsterMainPickupWeapon547210(unit, update, runtime)
}
