package server

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterHuntAggressionBits5449D0 = uint32(0x3f547ae1)

// monsterActionHunt5449D0 restores GAME.EXE 005449D0 without passing the
// native-width monster pointer through its original PE32 callback ABI. The
// action dispatcher ignores the original pointer-valued return.
func monsterActionHunt5449D0(unit *Object, push func(*Object, ai.ActionType) *AIStackItem) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || push == nil {
		return false
	}
	unit.UpdateDataMonster().Aggression = math.Float32frombits(monsterHuntAggressionBits5449D0)
	if item := push(unit, ai.ACTION_ROAM); item != nil {
		item.Args[0] = 0
		item.Args[2] = item.Args[2]&^uintptr(0xff) | uintptr(0x80)
	}
	return true
}

// MonsterActionHunt5449D0 binds the restored hunt update to the native action
// stack.
func (s *Server) MonsterActionHunt5449D0(unit *Object) bool {
	return monsterActionHunt5449D0(unit, func(unit *Object, action ai.ActionType) *AIStackItem {
		return unit.MonsterPushAction(action)
	})
}
