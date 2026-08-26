package server

import (
	"math"

	"github.com/opennox/libs/object"
)

type elevatorShaftCollisionHooks551C40 struct {
	typeIndex func(string) int
	link      func(*Object) *Object
}

// elevatorShaftCollision551C40 restores GAME.EXE 00551C40 without reading the
// linked elevator through ElevatorShaftUpdateData's PE32 pointer slot.
func elevatorShaftCollision551C40(shaft, candidate *Object, hooks elevatorShaftCollisionHooks551C40) {
	if shaft == nil || candidate == nil || shaft.UpdateData == nil {
		return
	}
	elevator := hooks.link(shaft)
	if elevator == nil || elevator.UpdateData == nil {
		return
	}
	for _, id := range [...]string{"SmallFist", "MediumFist", "LargeFist", "Meteor"} {
		if int(candidate.TypeInd) == hooks.typeIndex(id) {
			return
		}
	}
	switch candidate.Shape.Kind {
	case ShapeKindBox:
		if shaft.Shape.Box.W < candidate.Shape.Box.W || shaft.Shape.Box.H < candidate.Shape.Box.H {
			return
		}
	case ShapeKindCircle:
		diameter := candidate.Shape.Circle.R * 2
		if diameter > shaft.Shape.Box.W || diameter > shaft.Shape.Box.H {
			return
		}
	}
	if !MapPointInBox57B850(&shaft.NewPos, &shaft.Shape, &candidate.NewPos) {
		return
	}
	height := float32(int32(elevator.UpdateDataElevator().Field_4) - 64)
	if math.Abs(float64(candidate.ZVal-height)) > 10 {
		if height <= -10 {
			candidate.ObjFlags |= object.FlagInHole
			candidate.Pos39 = shaft.NewPos
			candidate.Field41 = math.Float32bits(elevator.NewPos.X)
			candidate.Field42 = math.Float32bits(elevator.NewPos.Y)
		}
		return
	}
	candidate.ObjFlags = candidate.ObjFlags&^object.FlagInHole | object.FlagOnObject
	candidate.Raise(height + 4)
	candidate.Field27 = 0
}

func (s *Server) ElevatorShaftCollision551C40(shaft, candidate *Object) {
	elevatorShaftCollision551C40(shaft, candidate, elevatorShaftCollisionHooks551C40{
		typeIndex: s.Types.IndByID,
		link:      (*Object).ElevatorLink,
	})
}
