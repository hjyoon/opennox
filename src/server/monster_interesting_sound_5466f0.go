package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterInterestingSoundHooks5466F0 struct {
	frame         func() uint32
	tickRate      func() uint32
	tileAt        func(types.Pointf) int
	pathReachable func(*Object, *types.Pointf) bool
	trace         func(types.Pointf, types.Pointf) bool
	random        func(int, int) int
	push          func(ai.ActionType, ...any) *AIStackItem
}

func monsterInterestingSound5466F0(unit *Object, hooks monsterInterestingSoundHooks5466F0) int {
	if unit == nil || unit.UpdateData == nil {
		return 0
	}
	update := unit.UpdateDataMonster()
	frame := hooks.frame()
	if update.Field97 == 0 || frame-update.Field101 >= 3*hooks.tickRate() {
		return 0
	}
	pos := types.Pointf{X: update.Field99X, Y: update.Field99Y}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterImmuneFire) || hooks.tileAt(pos) != 6 {
		if hooks.pathReachable(unit, &pos) {
			if hooks.trace(unit.PosVec, pos) {
				hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
				hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
				hooks.push(ai.ACTION_WAIT, frame+uint32(hooks.random(int(hooks.tickRate()), 2*int(hooks.tickRate()))))
				hooks.push(ai.ACTION_FACE_LOCATION, pos)
			} else {
				hooks.push(ai.DEPENDENCY_NO_INTERESTING_SOUND)
				hooks.push(ai.DEPENDENCY_NO_VISIBLE_ENEMY)
				hooks.push(ai.DEPENDENCY_NOT_FRUSTRATED)
				hooks.push(ai.DEPENDENCY_LOCATION_IS_SAFE, pos)
				hooks.push(ai.ACTION_MOVE_TO, pos, 0)
			}
		}
	}
	update.Field97 = 0
	return 1
}

// MonsterInterestingSound5466F0 restores the complete interesting-sound
// decision path while keeping Object, MonsterUpdateData, and AI stack pointers
// at native width. Tile and path callbacks are supplied by the top-level game
// server because those systems still live outside this package.
func (s *Server) MonsterInterestingSound5466F0(unit *Object, tileAt func(types.Pointf) int, pathReachable func(*Object, *types.Pointf) bool) int {
	return monsterInterestingSound5466F0(unit, monsterInterestingSoundHooks5466F0{
		frame:         s.Frame,
		tickRate:      s.TickRate,
		tileAt:        tileAt,
		pathReachable: pathReachable,
		trace: func(from, to types.Pointf) bool {
			return s.MapTraceRayAt(from, to, nil, nil, 9)
		},
		random: s.Rand.Logic.IntClamp,
		push:   unit.MonsterPushAction,
	})
}
