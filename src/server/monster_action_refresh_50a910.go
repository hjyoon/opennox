package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Each bit identifies an object-valued action argument in the original
// 005BF3F4 action metadata table. An action parameter occupies two PE32 words,
// so parameter 0 maps to Args[0] and parameter 1 maps to Args[2].
var monsterActionObjectArgMask50A910 = [72]uint8{
	ai.ACTION_ESCORT:                         1 << 1,
	ai.ACTION_MOVE_TO:                        1 << 1,
	ai.ACTION_FAR_MOVE_TO:                    1 << 1,
	ai.ACTION_MISSILE_ATTACK:                 1 << 1,
	ai.ACTION_CAST_SPELL_ON_OBJECT:           1 << 1,
	ai.ACTION_CAST_DURATION_SPELL:            1 << 1,
	ai.ACTION_FLEE:                           1 << 1,
	ai.ACTION_FACE_OBJECT:                    1 << 0,
	ai.ACTION_MOVE_TO_HOME:                   1 << 1,
	ai.DEPENDENCY_ALIVE:                      1 << 0,
	ai.DEPENDENCY_CAN_SEE:                    1 << 0,
	ai.DEPENDENCY_CANNOT_SEE:                 1 << 0,
	ai.DEPENDENCY_BLOCKED_LINE_OF_FIRE:       1 << 0,
	ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION: 1 << 1,
	ai.DEPENDENCY_OBJECT_FARTHER_THAN:        1 << 1,
	ai.DEPENDENCY_OBJECT_CLOSER_THAN:         1 << 1,
	ai.DEPENDENCY_NO_NEW_ENEMY:               1 << 0,
}

func monsterActionArgObject50A910(item *AIStackItem, param int) *Object {
	return (*Object)(unsafe.Pointer(item.Args[2*param]))
}

func monsterActionSetPos50A910(item *AIStackItem, target *Object) {
	item.Args[0] = uintptr(math.Float32bits(target.PosVec.X))
	item.Args[1] = uintptr(math.Float32bits(target.PosVec.Y))
}

func monsterActionRefresh50A910(unit *Object, canInteract func(*Object, *Object, int) bool) int {
	update := unit.UpdateDataMonster()
	if target := update.PreferredEnemy; target != nil && target.ObjFlags.HasAny(object.FlagDestroyed|object.FlagDead) {
		update.PreferredEnemy = nil
	}
	if update.AIStackInd < 0 {
		return int(update.AIStackInd)
	}
	for i := int(update.AIStackInd); i >= 0; i-- {
		item := &update.AIStack[i]
		if item.Action < uint32(len(monsterActionObjectArgMask50A910)) {
			mask := monsterActionObjectArgMask50A910[item.Action]
			for param := 0; param < 2; param++ {
				if mask&(1<<param) == 0 {
					continue
				}
				target := monsterActionArgObject50A910(item, param)
				if target != nil && target.ObjFlags.Has(object.FlagDestroyed) {
					item.Args[2*param] = 0
				}
			}
		}

		switch ai.ActionType(item.Action) {
		case ai.ACTION_ESCORT:
			if target := monsterActionArgObject50A910(item, 1); target != nil {
				monsterActionSetPos50A910(item, target)
			}
		case ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO:
			if target := monsterActionArgObject50A910(item, 1); target != nil {
				if canInteract(unit, target, 0) || update.HasAction(ai.ACTION_ESCORT) {
					monsterActionSetPos50A910(item, target)
				} else {
					item.Args[2] = 0
				}
			}
		case ai.ACTION_FIGHT:
			if target := update.CurrentEnemy; target != nil {
				monsterActionSetPos50A910(item, target)
			}
		case ai.ACTION_MISSILE_ATTACK:
			if target := monsterActionArgObject50A910(item, 1); target != nil && canInteract(unit, target, 0) {
				monsterActionSetPos50A910(item, target)
			}
		}
	}
	return 0
}

// MonsterActionRefresh50A910 updates pointer-bearing action arguments without
// interpreting MonsterUpdateData or AIStackItem through PE32 byte offsets.
func (s *Server) MonsterActionRefresh50A910(unit *Object) int {
	if unit == nil || unit.UpdateData == nil {
		return -1
	}
	return monsterActionRefresh50A910(unit, s.CanInteract)
}
