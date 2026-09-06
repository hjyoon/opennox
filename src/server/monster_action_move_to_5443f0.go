package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionMoveToHooks5443F0 struct {
	frame       func() uint32
	tickRate    func() uint32
	random      func(int, int) int
	setMovePath func(*Object, types.Pointf) bool
	pathReset   func() bool
	moveAudio   func(*Object)
	push        func(ai.ActionType, ...any) *AIStackItem
	pop         func() int
}

func monsterActionPrevious5443F0(update *MonsterUpdateData) ai.ActionType {
	for i := int(update.AIStackInd) - 1; i >= 0; i-- {
		if action := update.AIStack[i].Type(); !action.IsCondition() {
			return action
		}
	}
	return ai.ACTION_INVALID
}

func monsterActionMoveToForAction5443F0(unit *Object, action ai.ActionType, hooks monsterActionMoveToHooks5443F0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.tickRate == nil || hooks.random == nil ||
		hooks.setMovePath == nil || hooks.push == nil || hooks.pop == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != action {
		return false
	}
	if unit.SpeedBase < 0.0099999998 {
		hooks.pop()
		return true
	}
	target := head.ArgPos(0)
	if monsterActionPrevious5443F0(update) == ai.ACTION_ESCORT {
		runDistance := update.Field329 * 3
		delta := target.Sub(unit.PosVec)
		distance := float32(math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y)))
		if distance < runDistance {
			if !update.StatusFlags.Has(object.MonStatusAlwaysRun) {
				update.StatusFlags &^= object.MonStatusRunning
			}
		} else if distance > runDistance+30 && !update.StatusFlags.Has(object.MonStatusNeverRun) {
			update.StatusFlags |= object.MonStatusRunning
		}
	}
	if hooks.setMovePath(unit, target) {
		status := byte(update.Field71)
		retry := status == 2 || status == 1 && hooks.frame()-update.Field135 < 5*hooks.tickRate()
		if status == 1 {
			update.Field135 = hooks.frame()
		}
		pathReset := hooks.pathReset != nil && hooks.pathReset()
		if status == 0 && !pathReset && head.ArgObj(2) == nil {
			unit.Direction2 = DirFromVec(target.Sub(unit.PosVec))
			hooks.pop()
		}
		if retry {
			hooks.push(ai.DEPENDENCY_TIME, hooks.frame()+uint32(hooks.random(2*int(hooks.tickRate()), 4*int(hooks.tickRate()))))
			hooks.push(ai.ACTION_RANDOM_WALK)
			update.StatusFlags |= object.MonStatusFrustrated
		}
		if status != 0 {
			hooks.push(ai.ACTION_WAIT, hooks.frame()+uint32(hooks.random(int(hooks.tickRate()/2), int(hooks.tickRate()))))
		}
	}
	if hooks.moveAudio != nil {
		hooks.moveAudio(unit)
	}
	return true
}

func monsterActionMoveTo5443F0(unit *Object, hooks monsterActionMoveToHooks5443F0) bool {
	return monsterActionMoveToForAction5443F0(unit, ai.ACTION_MOVE_TO, hooks)
}

// monsterActionMoveToHome544950 is the original 00544950 wrapper expressed
// without its PE32 int parameter. Its update body is exactly 005443F0, but the
// dispatch admission remains ACTION_MOVE_TO_HOME.
func monsterActionMoveToHome544950(unit *Object, hooks monsterActionMoveToHooks5443F0) bool {
	return monsterActionMoveToForAction5443F0(unit, ai.ACTION_MOVE_TO_HOME, hooks)
}

// monsterCreatureSetMovePath50D5A0 is the native-width movement core needed
// by 005443F0. Detailed path generation already owns obstacle and waypoint
// routing in the Go server; this adapter preserves the original eight-unit
// arrival boundary, path refresh cadence, and movement completion result.
func (s *Server) monsterCreatureSetMovePath50D5A0(unit *Object, target types.Pointf, setDetailedPath func(*Object, *types.Pointf)) bool {
	update := unit.UpdateDataMonster()
	delta := target.Sub(unit.PosVec)
	if math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y))+0.000099999997 <= 8.0 {
		return true
	}
	lastDelta := update.Field68.Sub(target)
	if update.Field2 == 0 || s.Frame()-update.Field70 > 10 && lastDelta.X*lastDelta.X+lastDelta.Y*lastDelta.Y > 2500.0 {
		if setDetailedPath != nil {
			setDetailedPath(unit, &target)
		}
	}
	if update.Field2 != 0 && monsterCreatureActuallyMove50D3B0(unit, s.MapTraceRay) {
		update.Field2 = 0
		return true
	}
	return false
}

// MonsterActionMoveTo5443F0 binds GAME.EXE 005443F0 to native Object,
// MonsterUpdateData, and AI-stack pointers.
func (s *Server) monsterActionMoveToHooks5443F0(unit *Object, setDetailedPath func(*Object, *types.Pointf)) monsterActionMoveToHooks5443F0 {
	return monsterActionMoveToHooks5443F0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		random:   s.Rand.Logic.IntClamp,
		setMovePath: func(unit *Object, target types.Pointf) bool {
			return s.monsterCreatureSetMovePath50D5A0(unit, target, setDetailedPath)
		},
		pathReset: func() bool { return false },
		moveAudio: s.monsterMoveAudio534030,
		push:      unit.MonsterPushAction,
		pop:       unit.MonsterPopAction,
	}
}

func (s *Server) MonsterActionMoveTo5443F0(unit *Object, setDetailedPath func(*Object, *types.Pointf)) bool {
	if unit == nil {
		return false
	}
	return monsterActionMoveTo5443F0(unit, s.monsterActionMoveToHooks5443F0(unit, setDetailedPath))
}

// MonsterActionMoveToHome544950 binds the ACTION_MOVE_TO_HOME update wrapper
// to native-width server state. Its start, end, and cancel callbacks reuse the
// run-state helpers at 00534750 and 00534780.
func (s *Server) MonsterActionMoveToHome544950(unit *Object, setDetailedPath func(*Object, *types.Pointf)) bool {
	if unit == nil {
		return false
	}
	return monsterActionMoveToHome544950(unit, s.monsterActionMoveToHooks5443F0(unit, setDetailedPath))
}
