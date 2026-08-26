package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterFacingDirection534120(unit *Object, direction types.Pointf) bool {
	forward := unit.Direction1.Vec()
	return float64(forward.X)*float64(direction.X)+float64(forward.Y)*float64(direction.Y) > 0.89999998
}

func monsterTurnTowardDirection545240(unit *Object, direction types.Pointf) {
	forward := unit.Direction1.Vec()
	delta := 8
	if float64(direction.Y)*float64(forward.X)-float64(direction.X)*float64(forward.Y) < 0 {
		delta = -8
	}
	unit.Direction2 = RoundDir(int(unit.Direction2) + delta)
}

// monsterActionFacePoint545240 restores GAME.EXE 00545240. The original
// routine turns Direction2 by one 8/256 revolution step and completes the
// current action once Direction1 is within the original 0.89999998 dot-product
// threshold. Keeping the point and stack access in Go avoids interpreting a
// native-width Object through the PE32 offsets used by GAME.EXE.
func monsterActionFacePoint545240(unit *Object, target types.Pointf, pop func() int) int {
	dx := float64(target.X - unit.PosVec.X)
	dy := float64(target.Y - unit.PosVec.Y)
	distance := math.Sqrt(dx*dx+dy*dy) + 0.001
	direction := types.Ptf(float32(dx/distance), float32(dy/distance))
	monsterTurnTowardDirection545240(unit, direction)
	if monsterFacingDirection534120(unit, direction) {
		return pop()
	}
	return 0
}

func monsterFaceActionHead545210(unit *Object, action ai.ActionType) *AIStackItem {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return nil
	}
	head := unit.UpdateDataMonster().AIStackHead()
	if head == nil || head.Type() != action {
		return nil
	}
	return head
}

func monsterActionFaceLocation545210(unit *Object, pop func() int) int {
	head := monsterFaceActionHead545210(unit, ai.ACTION_FACE_LOCATION)
	if head == nil {
		return 0
	}
	return monsterActionFacePoint545240(unit, head.ArgPos(0), pop)
}

func monsterActionFaceObject545300(unit *Object, pop func() int) int {
	head := monsterFaceActionHead545210(unit, ai.ACTION_FACE_OBJECT)
	if head == nil {
		return 0
	}
	target := head.ArgObj(0)
	if target == nil {
		return pop()
	}
	return monsterActionFacePoint545240(unit, target.PosVec, pop)
}

func monsterActionFaceAngle545340(unit *Object, pop func() int) int {
	head := monsterFaceActionHead545210(unit, ai.ACTION_FACE_ANGLE)
	if head == nil {
		return 0
	}
	direction := Dir16(head.ArgU32(0)).Vec()
	monsterTurnTowardDirection545240(unit, direction)
	if monsterFacingDirection534120(unit, direction) {
		return pop()
	}
	return 0
}

func monsterActionSetAngle5453E0(unit *Object, pop func() int) int {
	head := monsterFaceActionHead545210(unit, ai.ACTION_SET_ANGLE)
	if head == nil {
		return 0
	}
	unit.Direction2 = RoundDir(int(int32(head.ArgU32(0))))
	unit.Direction1 = unit.Direction2
	return pop()
}

func (s *Server) MonsterActionFaceLocation545210(unit *Object) int {
	return monsterActionFaceLocation545210(unit, unit.MonsterPopAction)
}

func (s *Server) MonsterActionFaceObject545300(unit *Object) int {
	return monsterActionFaceObject545300(unit, unit.MonsterPopAction)
}

func (s *Server) MonsterActionFaceAngle545340(unit *Object) int {
	return monsterActionFaceAngle545340(unit, unit.MonsterPopAction)
}

func (s *Server) MonsterActionSetAngle5453E0(unit *Object) int {
	return monsterActionSetAngle5453E0(unit, unit.MonsterPopAction)
}
