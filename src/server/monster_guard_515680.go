package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterGoPatrol515680 restores GAME.EXE 00515680 with native Object and
// MonsterUpdateData pointers. The script builtin calls this routine "Guard";
// p2 is used only to derive the initial facing at p1.
func (s *Server) MonsterGoPatrol515680(unit *Object, p1, p2 types.Pointf, distance float32) bool {
	if unit == nil || unit.UpdateData == nil ||
		!unit.ObjClass.Has(object.ClassMonster) || unit.ObjFlags.Has(object.FlagDead) {
		return false
	}
	update := unit.UpdateDataMonster()
	unit.ClearActionStack()
	unit.MonsterPushAction(ai.ACTION_GUARD, p1, uint32(DirFromVec(types.Ptf(p2.X-p1.X, p2.Y-p1.Y))))
	update.SightRange = distance
	return true
}
