package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
)

type elevatorShaftCollisionHooks551C40 struct {
	typeIndex func(string) int
	link      func(*Object, unsafe.Pointer) *Object
	raise     func(*Object, float32)
}

// elevatorShaftCollision551C40 restores GAME.EXE 00551C40 without reading the
// linked elevator through ElevatorShaftUpdateData's PE32 pointer slot.
func elevatorShaftCollision551C40(shaft, candidate *Object, hooks elevatorShaftCollisionHooks551C40) {
	if shaft == nil || candidate == nil || shaft.UpdateData == nil {
		return
	}
	data := shaft.UpdateData
	ids := elevatorCollisionTypes551BF0(hooks.typeIndex)
	elevator := hooks.link(shaft, data)
	if elevator == nil || elevator.UpdateData == nil {
		return
	}
	for _, id := range ids {
		if int(candidate.TypeInd) == id {
			return
		}
	}
	switch candidate.Shape.Kind {
	case ShapeKindBox:
		if !(shaft.Shape.Box.W >= candidate.Shape.Box.W) || !(shaft.Shape.Box.H >= candidate.Shape.Box.H) {
			return
		}
	case ShapeKindCircle:
		radius := float64(candidate.Shape.Circle.R)
		diameter := objectForceAdd53_52DF80(radius, radius)
		if diameter > float64(shaft.Shape.Box.W) || diameter > float64(shaft.Shape.Box.H) {
			return
		}
	}
	if !MapPointInBox57B850(&shaft.NewPos, &shaft.Shape, &candidate.NewPos) {
		return
	}
	height := experienceSpill32TowardZero(float64(int32(elevator.UpdateDataElevator().Field_4 - 64)))
	difference := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(candidate.ZVal), objectForceMemory32_52DF80(height)))
	if math.Abs(float64(difference)) > 10 {
		if height <= -10 {
			candidate.ObjFlags |= object.FlagInHole
			candidate.Pos39 = shaft.NewPos
			candidate.Field41 = math.Float32bits(elevator.NewPos.X)
			candidate.Field42 = math.Float32bits(elevator.NewPos.Y)
		}
		return
	}
	candidate.ObjFlags = candidate.ObjFlags&^object.FlagInHole | object.FlagOnObject
	hooks.raise(candidate, experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(height), 4)))
	candidate.Field27 = 0
}

func (s *Server) ElevatorShaftCollision551C40(shaft, candidate *Object, typeIndex func(string) int) {
	elevatorShaftCollision551C40(shaft, candidate, elevatorShaftCollisionHooks551C40{
		typeIndex: typeIndex,
		link:      (*Object).ElevatorLinkFor,
		raise:     (*Object).Raise,
	})
}
