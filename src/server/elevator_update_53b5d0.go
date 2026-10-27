package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type elevatorUpdateHooks53B5D0 struct {
	frame          func() uint32
	tickRate       func() uint32
	link           func(*Object, unsafe.Pointer) *Object
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

func elevatorCarryUp53B750(elevator *Object, hooks elevatorUpdateHooks53B5D0) {
	hooks.eachInCircle(elevator.PosVec, 64, func(candidate *Object) bool {
		elevatorPassengerUp53B750(candidate, elevator, hooks)
		return true
	})
}

func elevatorPassengerUp53B750(candidate, elevator *Object, hooks elevatorUpdateHooks53B5D0) {
	if candidate == nil || !hooks.pointInBox(&elevator.PosVec, &elevator.Shape, &candidate.PosVec) {
		return
	}
	data := elevator.UpdateDataElevator()
	shaft := hooks.link(elevator, unsafe.Pointer(data))
	if shaft == nil {
		return
	}
	difference := experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(candidate.ZVal), float64(int32(data.Field_4))))
	if math.Abs(float64(difference)) >= 10 {
		return
	}
	fit := true
	switch candidate.Shape.Kind {
	case ShapeKindBox:
		fit = shaft.Shape.Box.W >= candidate.Shape.Box.W && shaft.Shape.Box.H >= candidate.Shape.Box.H
	case ShapeKindCircle:
		radius := float64(candidate.Shape.Circle.R)
		diameter := objectForceAdd53_52DF80(radius, radius)
		fit = !(diameter > float64(shaft.Shape.Box.W)) && !(diameter > float64(shaft.Shape.Box.H))
	}
	if !fit {
		candidate.ObjFlags &^= object.FlagOnObject
		hooks.raise(candidate, 0)
		return
	}
	// Destination and height are live loads through the callback's cached data.
	hooks.move(candidate, hooks.link(elevator, unsafe.Pointer(data)).PosVec)
	hooks.raise(candidate, experienceSpill32TowardZero(float64(int32(data.Field_4-64))))
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
		if int32(update.Field_4) > 0 {
			update.Field_4 -= 2
		} else {
			update.Field_3 = 0
			unit.Field34 = hooks.frame()
		}
		hooks.needSync(unit)
		if shaft := hooks.link(unit, unsafe.Pointer(update)); shaft != nil {
			hooks.needSync(shaft)
		}
		if int32(update.Field_4) <= 20 {
			unit.ObjFlags |= object.FlagShort
		}
	case 2:
		if unit.ObjFlags.Has(object.FlagEnabled) && hooks.frame()-unit.Field34 > hooks.tickRate() {
			update.Field_3 = 1
			hooks.audio(unit, false)
		}
	case 3:
		update.Field_4 += 2
		hooks.needSync(unit)
		if shaft := hooks.link(unit, unsafe.Pointer(update)); shaft != nil {
			hooks.needSync(shaft)
		}
		if int32(update.Field_4) >= 20 {
			unit.ObjFlags &^= object.FlagShort
		}
		if int32(update.Field_4) >= 32 {
			elevatorCarryUp53B750(unit, hooks)
		}
		if int32(update.Field_4) >= 64 {
			update.Field_4 = 64
			update.Field_3 = 2
			unit.Field34 = hooks.frame()
		}
	}
}

func elevatorShaftCarryDown53B410(shaft *Object, hooks elevatorUpdateHooks53B5D0) {
	hooks.eachInCircle(shaft.PosVec, 64, func(candidate *Object) bool {
		elevatorPassengerDown53B410(candidate, shaft, hooks)
		return true
	})
}

func elevatorPassengerDown53B410(candidate, shaft *Object, hooks elevatorUpdateHooks53B5D0) {
	if candidate == nil || !hooks.pointInBox(&shaft.PosVec, &shaft.Shape, &candidate.PosVec) {
		return
	}
	data := shaft.UpdateData
	elevator := hooks.link(shaft, data)
	update := elevator.UpdateDataElevator()
	level := objectForceSub53_52DF80(objectForceAdd53_52DF80(float64(candidate.ZVal), 64), float64(int32(update.Field_4)))
	difference := experienceSpill32TowardZero(level)
	if math.Abs(float64(difference)) >= 10 {
		return
	}
	hooks.move(candidate, hooks.link(shaft, data).PosVec)
	hooks.raise(candidate, experienceSpill32TowardZero(float64(int32(update.Field_4))))
}

// elevatorShaftUpdate53B380 restores GAME.EXE 0053B380 and its passenger
// callback without loading the linked elevator from a 32-bit pointer slot.
func elevatorShaftUpdate53B380(unit *Object, hooks elevatorUpdateHooks53B5D0) uint8 {
	if unit == nil || unit.UpdateData == nil {
		return 0
	}
	update := unit.UpdateDataElevatorShaft()
	if hooks.link(unit, unsafe.Pointer(update)) == nil {
		return 0
	}
	hooks.queueCollision(unit)
	elevator := hooks.link(unit, unsafe.Pointer(update))
	elevatorUpdate := elevator.UpdateDataElevator()
	state := elevatorUpdate.Field_3
	if state == 1 {
		if int32(elevatorUpdate.Field_4) <= 32 {
			elevatorShaftCarryDown53B410(unit, hooks)
		}
		if update.Field_3 != elevatorUpdate.Field_3 {
			hooks.audio(unit, false)
		}
	} else if state == 3 && update.Field_3 != state {
		hooks.audio(unit, true)
	}
	update.Field_3 = elevatorUpdate.Field_3
	return update.Field_3
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
		link:     (*Object).ElevatorLinkFor,
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

func (s *Server) ElevatorShaftUpdate53B380(unit *Object, runtime ElevatorUpdateRuntime53B5D0) uint8 {
	return elevatorShaftUpdate53B380(unit, s.elevatorHooks53B5D0(runtime))
}

func (s *Server) ElevatorPassengerUp53B750(candidate, elevator *Object, runtime ElevatorUpdateRuntime53B5D0) {
	elevatorPassengerUp53B750(candidate, elevator, s.elevatorHooks53B5D0(runtime))
}

func (s *Server) ElevatorPassengerDown53B410(candidate, shaft *Object, runtime ElevatorUpdateRuntime53B5D0) {
	elevatorPassengerDown53B410(candidate, shaft, s.elevatorHooks53B5D0(runtime))
}

func (s *Server) ElevatorAudio53B490(unit *Object, upward bool) {
	s.elevatorHooks53B5D0(ElevatorUpdateRuntime53B5D0{}).audio(unit, upward)
}
