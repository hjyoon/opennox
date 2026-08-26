package server

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterMainPassiveShopkeeper547210 handles the exact no-op path used by a
// passive shopkeeper in a regular game. The original function reaches no
// state-changing branch when the shopkeeper is idle, healthy, unbuffed,
// unarmed, has no target, and has no monster status capabilities enabled.
//
// It returns false for every other state so that callers cannot accidentally
// treat this narrow, verified path as the complete 00547210 implementation.
func (s *Server) MonsterMainPassiveShopkeeper547210(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil ||
		!unit.ObjClass.Has(object.ClassMonster) ||
		!unit.ObjSubClass.AsMonster().Has(object.MonsterShopkeeper) ||
		noxflags.HasGame(noxflags.GameModeQuest) ||
		unit.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		unit.Buffs != 0 || unit.InvFirstItem != nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_IDLE ||
		update.CurrentEnemy != nil || update.StatusFlags != 0 {
		return false
	}
	if health := unit.HealthData; health != nil && health.Max != 0 && health.Cur < health.Max {
		return false
	}
	return true
}
