package server

import (
	"math"

	"github.com/opennox/libs/object"
)

type elevatorCollisionHooks551AE0 struct {
	typeIndex func(string) int
	circleBox func(*Object, *Object, bool)
	boxBox    func(*Object, *Object)
	raise     func(*Object, float32)
}

// elevatorCollision551AE0 restores GAME.EXE 00551AE0 without truncating the
// elevator, passenger, or ElevatorUpdateData pointers to PE32 integers.
func elevatorCollision551AE0(elevator, candidate *Object, candidateMoves bool, hooks elevatorCollisionHooks551AE0) {
	if elevator == nil || candidate == nil || elevator.UpdateData == nil {
		return
	}
	update := elevator.UpdateDataElevator()
	ids := elevatorCollisionTypes551BF0(hooks.typeIndex)
	if !candidateMoves {
		return
	}
	for _, id := range ids {
		if int(candidate.TypeInd) == id {
			return
		}
	}
	height := float64(int32(update.Field_4))
	difference := experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(candidate.ZVal), height))
	if math.Abs(float64(difference)) > 10 {
		if height > float64(candidate.ZVal) {
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
	hooks.raise(candidate, experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(int32(update.Field_4)), 4)))
	candidate.Field27 = 0
}

func elevatorCollisionTypes551BF0(read func(string) int) [4]int {
	return [4]int{read("SmallFist"), read("MediumFist"), read("LargeFist"), read("Meteor")}
}

func (s *Server) ElevatorCollision551AE0(elevator, candidate *Object, candidateMoves bool, typeIndex func(string) int, circleBox func(*Object, *Object, bool), boxBox func(*Object, *Object)) {
	elevatorCollision551AE0(elevator, candidate, candidateMoves, elevatorCollisionHooks551AE0{
		typeIndex: typeIndex,
		circleBox: circleBox,
		boxBox:    boxBox,
		raise:     (*Object).Raise,
	})
}
