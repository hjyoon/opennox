package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// PlayerAttackTargetRuntime5386A0 keeps allegiance, visibility and box
// collision as services. Candidate/attacker pointers are native-width;
// selection does not change health, clocks, animation or a gameplay RNG.
type PlayerAttackTargetRuntime5386A0 struct {
	IsEnemy     func(*Object, *Object) bool
	CanInteract func(*Object, *Object, int) bool
	Direction   func(Dir16) types.Pointf
	Box         func(types.Pointf, float32, *Object) (float64, types.Pointf)
}

// PlayerAttackSelectCandidate5386A0 models the original 005386A0 callback.
// Precision 53 / ToZero is explicit, without changing the host thread's FPU.
// The original first monster is sticky; a monster can replace a closer prop.
func PlayerAttackSelectCandidate5386A0(candidate, attacker *Object, bestRange *float32, closest **Object, h PlayerAttackTargetRuntime5386A0) {
	if candidate == nil || attacker == nil || candidate == attacker || bestRange == nil || closest == nil {
		return
	}
	flags := candidate.ObjFlags
	if flags&0x8049 != 0 || candidate.ObjClass&6 == 0 && flags&0x10 != 0 && flags&0x80 == 0 {
		return
	}
	// 005386E7 calls allegiance even for props. Its false branch reloads
	// class, and 005386FD reloads destroyed flags after the callback.
	if !h.IsEnemy(attacker, candidate) && candidate.ObjClass&6 != 0 {
		return
	}
	if candidate.ObjFlags&0x8000 != 0 || !h.CanInteract(attacker, candidate, 1) || !(*bestRange > 0) {
		return
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	div := pointOnLineDivChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	// X is FSTP binary32; Y is FST binary32 but remains retained for
	// multiplication by the spilled Y. FSQRT and the cone dot stay retained.
	dx := float64(spill(add(float64(candidate.PosVec.X), -float64(attacker.PosVec.X))))
	dy := add(float64(candidate.PosVec.Y), -float64(attacker.PosVec.Y))
	sy := float64(spill(dy))
	distance := monsterMoveForceSqrtChop53_50D50C(add(mul(dy, sy), mul(dx, dx)))
	if !(distance > 0) {
		return
	}
	facing := h.Direction(attacker.Direction1)
	dot := add(mul(div(sy, distance), float64(facing.Y)), mul(div(dx, distance), float64(facing.X)))
	if !(dot > 0.5) {
		return
	}
	switch candidate.Shape.Kind {
	case ShapeKindCircle:
		distance = add(distance, -float64(candidate.Shape.Circle.R))
	case ShapeKindBox:
		edge, _ := h.Box(attacker.PosVec, *bestRange, candidate)
		// C0 rejects both negative and unordered box returns. The return
		// stays binary64 until the original final FSTP, not a float32 cast.
		if !(edge >= 0) {
			return
		}
		distance = add(float64(*bestRange), -edge)
	}
	// The original C0 test also clamps an unordered circle edge to zero.
	if !(distance >= 0) {
		distance = 0
	}
	old := *closest
	if (!(distance >= float64(*bestRange)) || old != nil && old.ObjClass&object.ClassMonster == 0 && candidate.ObjClass&object.ClassMonster != 0) &&
		(old == nil || old.ObjClass&object.ClassMonster == 0) {
		*bestRange, *closest = spill(distance), candidate
	}
}
