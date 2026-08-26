package server

import (
	"math"

	"github.com/opennox/libs/object"
)

type elevatorCollisionHooks551AE0 struct {
	typeIndex func(string) int
	circleBox func(*Object, *Object, bool)
	boxBox    func(*Object, *Object)
}

// elevatorCollision551AE0 restores GAME.EXE 00551AE0 without truncating the
// elevator, passenger, or ElevatorUpdateData pointers to PE32 integers.
func elevatorCollision551AE0(elevator, candidate *Object, candidateMoves bool, hooks elevatorCollisionHooks551AE0) {
	if elevator == nil || candidate == nil || elevator.UpdateData == nil || !candidateMoves {
		return
	}
	for _, id := range [...]string{"SmallFist", "MediumFist", "LargeFist", "Meteor"} {
		if int(candidate.TypeInd) == hooks.typeIndex(id) {
			return
		}
	}
	height := float32(int32(elevator.UpdateDataElevator().Field_4))
	if math.Abs(float64(candidate.ZVal-height)) > 10 {
		if height > candidate.ZVal {
			switch candidate.Shape.Kind {
			case ShapeKindCircle:
				hooks.circleBox(candidate, elevator, false)
			case ShapeKindBox:
				hooks.boxBox(candidate, elevator)
			}
		}
		return
	}
	if !MapPointInBox57B850(&elevator.NewPos, &elevator.Shape, &candidate.NewPos) {
		return
	}
	candidate.ObjFlags = candidate.ObjFlags&^object.FlagInHole | object.FlagOnObject
	candidate.Raise(height + 4)
	candidate.Field27 = 0
}

func (s *Server) ElevatorCollision551AE0(elevator, candidate *Object, candidateMoves bool, circleBox func(*Object, *Object, bool), boxBox func(*Object, *Object)) {
	elevatorCollision551AE0(elevator, candidate, candidateMoves, elevatorCollisionHooks551AE0{
		typeIndex: s.Types.IndByID,
		circleBox: circleBox,
		boxBox:    boxBox,
	})
}
