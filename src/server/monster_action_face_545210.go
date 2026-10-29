package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterFacingDirection534120(unit *Object, direction types.Pointf) bool {
	forward := unit.Direction1.Vec()
	y := objectForceMul53_52DF80(objectForceMemory32_52DF80(forward.Y), objectForceMemory32_52DF80(direction.Y))
	x := objectForceMul53_52DF80(objectForceMemory32_52DF80(forward.X), objectForceMemory32_52DF80(direction.X))
	return objectForceAdd53_52DF80(y, x) > float64(math.Float32frombits(0x3f666666))
}

func monsterTurnTowardDirection545240(unit *Object, direction types.Pointf, retainedY float64) {
	forward := unit.Direction1.Vec()
	y := objectForceMul53_52DF80(retainedY, objectForceMemory32_52DF80(forward.X))
	x := objectForceMul53_52DF80(objectForceMemory32_52DF80(direction.X), objectForceMemory32_52DF80(forward.Y))
	cross := objectForceSub53_52DF80(y, x)
	delta := 8
	// Original C0 selects the negative turn for unordered comparisons too.
	if !(cross >= 0) {
		delta = -8
	}
	unit.Direction2 = RoundDir(int(unit.Direction2) + delta)
}

// monsterActionFacePoint545240 restores GAME.EXE 00545240. The original
// routine turns Direction2 by one 8/256 revolution step and completes the
// current action once Direction1 is within the original 0.89999998 dot-product
// threshold. Retained PC53/chop values, F32 spills and memory constants are
// modeled locally without changing host FP state. The Y quotient remains
// unspilled for the cross product, while the facing predicate reads its spill.
func monsterActionFacePoint545240(unit *Object, target types.Pointf, pop func() int) int {
	dx := objectForceSub53_52DF80(objectForceMemory32_52DF80(target.X), objectForceMemory32_52DF80(unit.PosVec.X))
	dy := objectForceSub53_52DF80(objectForceMemory32_52DF80(target.Y), objectForceMemory32_52DF80(unit.PosVec.Y))
	storedY := experienceSpill32TowardZero(dy)
	squareY := objectForceMul53_52DF80(dy, objectForceMemory32_52DF80(storedY))
	squareX := objectForceMul53_52DF80(dx, dx)
	distance := monsterMoveForceSqrtChop53_50D50C(objectForceAdd53_52DF80(squareY, squareX))
	denominator := experienceSpill32TowardZero(objectForceAdd53_52DF80(distance, float64(math.Float32frombits(0x3a83126f))))
	dirX := experienceSpill32TowardZero(objectForceDiv53_52DF80(dx, objectForceMemory32_52DF80(denominator)))
	retainedY := objectForceDiv53_52DF80(objectForceMemory32_52DF80(storedY), objectForceMemory32_52DF80(denominator))
	direction := types.Ptf(dirX, experienceSpill32TowardZero(retainedY))
	monsterTurnTowardDirection545240(unit, direction, retainedY)
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
	monsterTurnTowardDirection545240(unit, direction, objectForceMemory32_52DF80(direction.Y))
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
