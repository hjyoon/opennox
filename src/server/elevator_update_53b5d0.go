package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type elevatorUpdateHooks53B5D0 struct {
	frame          func() uint32
	tickRate       func() uint32
	link           func(*Object) *Object
	needSync       func(*Object)
	audio          func(*Object, bool)
	eachInCircle   func(types.Pointf, float32, func(*Object) bool)
	pointInBox     func(*types.Pointf, *Shape, *types.Pointf) bool
	move           func(*Object, types.Pointf)
	raise          func(*Object, float32)
	queueCollision func(*Object)
}

func elevatorSound53B490(unit *Object, upward bool) sound.ID {
	if unit == nil {
		return 0
	}
	switch unit.Material {
	case 8:
		if upward {
			return 257
		}
		return 258
	case 16:
		if unit.ObjSubClass&0x20 != 0 {
			if upward {
				return 253
			}
			return 254
		}
		if unit.ObjSubClass&0x40 != 0 {
			if upward {
				return 259
			}
			return 260
		}
		if upward {
			return 251
		}
		return 252
	case 32:
		if unit.ObjSubClass&0x2 != 0 {
			if upward {
				return 255
			}
			return 256
		}
		if upward {
			return 249
		}
		return 250
	default:
		return 0
	}
}

func elevatorCarryUp53B750(elevator *Object, update *ElevatorUpdateData, hooks elevatorUpdateHooks53B5D0) {
	shaft := hooks.link(elevator)
	if shaft == nil {
		return
	}
	height := update.Field_4
	hooks.eachInCircle(elevator.PosVec, 64, func(candidate *Object) bool {
		if candidate == nil || !hooks.pointInBox(&elevator.PosVec, &elevator.Shape, &candidate.PosVec) {
			return true
		}
		if math.Abs(float64(candidate.ZVal)-float64(int32(height))) >= 10.0 {
			return true
		}
		switch candidate.Shape.Kind {
		case ShapeKindBox:
			if float64(shaft.Shape.Box.W) < float64(candidate.Shape.Box.W) ||
				float64(shaft.Shape.Box.H) < float64(candidate.Shape.Box.H) {
				candidate.ObjFlags &^= object.FlagOnObject
				hooks.raise(candidate, 0)
				return true
			}
		case ShapeKindCircle:
			diameter := float64(candidate.Shape.Circle.R) * 2.0
			if diameter > float64(shaft.Shape.Box.W) || diameter > float64(shaft.Shape.Box.H) {
				candidate.ObjFlags &^= object.FlagOnObject
				hooks.raise(candidate, 0)
				return true
			}
		}
		hooks.move(candidate, shaft.PosVec)
		hooks.raise(candidate, float32(int32(height-64)))
		return true
	})
}

// elevatorUpdate53B5D0 restores GAME.EXE 0053B5D0 with the original unsigned
// timers, signed height comparisons, synchronization order, and passenger
// callback boundary.
func elevatorUpdate53B5D0(unit *Object, hooks elevatorUpdateHooks53B5D0) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataElevator()
	switch update.Field_3 {
	case 0:
		if unit.ObjFlags.Has(object.FlagEnabled) && hooks.frame()-unit.Field34 > hooks.tickRate() {
			update.Field_3 = 3
			hooks.audio(unit, true)
		}
	case 1:
		if update.Field_4 > 0 {
			update.Field_4 -= 2
		} else {
			update.Field_3 = 0
			unit.Field34 = hooks.frame()
		}
		hooks.needSync(unit)
		if shaft := hooks.link(unit); shaft != nil {
			hooks.needSync(shaft)
		}
		if int32(update.Field_4) <= 20 {
			unit.ObjFlags |= object.FlagNoCollide
		}
	case 2:
		if unit.ObjFlags.Has(object.FlagEnabled) && hooks.frame()-unit.Field34 > hooks.tickRate() {
			update.Field_3 = 1
			hooks.audio(unit, false)
		}
	case 3:
		update.Field_4 += 2
		hooks.needSync(unit)
		if shaft := hooks.link(unit); shaft != nil {
			hooks.needSync(shaft)
		}
		if int32(update.Field_4) >= 20 {
			unit.ObjFlags &^= object.FlagNoCollide
		}
		if int32(update.Field_4) >= 32 {
			elevatorCarryUp53B750(unit, update, hooks)
		}
		if update.Field_4 >= 64 {
			update.Field_4 = 64
			update.Field_3 = 2
			unit.Field34 = hooks.frame()
		}
	}
}

func elevatorShaftCarryDown53B410(shaft, elevator *Object, height uint32, hooks elevatorUpdateHooks53B5D0) {
	hooks.eachInCircle(shaft.PosVec, 64, func(candidate *Object) bool {
		if candidate == nil || !hooks.pointInBox(&shaft.PosVec, &shaft.Shape, &candidate.PosVec) {
			return true
		}
		level := float64(candidate.ZVal) + 64.0 - float64(int32(height))
		if math.Abs(level) < 10.0 {
			hooks.move(candidate, elevator.PosVec)
			hooks.raise(candidate, float32(int32(height)))
		}
		return true
	})
}

// elevatorShaftUpdate53B380 restores GAME.EXE 0053B380 and its passenger
// callback without loading the linked elevator from a 32-bit pointer slot.
func elevatorShaftUpdate53B380(unit *Object, hooks elevatorUpdateHooks53B5D0) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	update := unit.UpdateDataElevatorShaft()
	elevator := hooks.link(unit)
	if elevator == nil {
		return
	}
	hooks.queueCollision(unit)
	elevatorUpdate := elevator.UpdateDataElevator()
	state := elevatorUpdate.Field_3
	if state == 1 {
		if elevatorUpdate.Field_4 <= 32 {
			elevatorShaftCarryDown53B410(unit, elevator, elevatorUpdate.Field_4, hooks)
		}
		if update.Field_3 != state {
			hooks.audio(unit, false)
		}
	} else if state == 3 && update.Field_3 != state {
		hooks.audio(unit, true)
		update.Field_3 = state
		return
	}
	update.Field_3 = state
}

// ElevatorUpdateRuntime53B5D0 contains the two services still owned by the
// top-level game package. All map iteration, fixed-record state, audio, and
// native link handling stay in server.
type ElevatorUpdateRuntime53B5D0 struct {
	Move           func(*Object, types.Pointf)
	QueueCollision func(*Object)
}

func (s *Server) elevatorHooks53B5D0(runtime ElevatorUpdateRuntime53B5D0) elevatorUpdateHooks53B5D0 {
	return elevatorUpdateHooks53B5D0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		link:     (*Object).ElevatorLink,
		needSync: (*Object).NeedSync,
		audio: func(unit *Object, upward bool) {
			if id := elevatorSound53B490(unit, upward); id != 0 {
				s.Audio.EventObj(id, unit, 0, 0)
			}
		},
		eachInCircle: s.Map.EachObjInCircle,
		pointInBox:   MapPointInBox57B850,
		move:         runtime.Move,
		raise: func(unit *Object, height float32) {
			unit.Raise(height)
		},
		queueCollision: runtime.QueueCollision,
	}
}

func (s *Server) ElevatorUpdate53B5D0(unit *Object, runtime ElevatorUpdateRuntime53B5D0) {
	elevatorUpdate53B5D0(unit, s.elevatorHooks53B5D0(runtime))
}

func (s *Server) ElevatorShaftUpdate53B380(unit *Object, runtime ElevatorUpdateRuntime53B5D0) {
	elevatorShaftUpdate53B380(unit, s.elevatorHooks53B5D0(runtime))
}
