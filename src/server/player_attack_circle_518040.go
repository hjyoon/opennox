package server

import "github.com/opennox/libs/types"

// PlayerAttackScanCircle518040 models the original 00518040 rectangle and
// 005180B0 shape filter. Enumeration, box collision and visits are services.
// The center remains live across callbacks, as in the original pointer
// context. Precision 53 / ToZero is explicit without changing the host FPU.
func PlayerAttackScanCircle518040(center *types.Pointf, radius float32, eachRect func(types.Rectf, func(*Object) bool), box func(types.Pointf, float32, *Object) (float64, types.Pointf), visit func(*Object)) {
	if center == nil || visit == nil {
		return
	}
	add := pointOnLineAddChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	r := float64(radius)
	rect := types.Rectf{
		Min: types.Ptf(spill(add(float64(center.X), -r)), spill(add(float64(center.Y), -r))),
		Max: types.Ptf(spill(add(r, float64(center.X))), spill(add(r, float64(center.Y)))),
	}
	eachRect(rect, func(candidate *Object) bool {
		if PlayerAttackCircleOverlaps5180B0(center, radius, candidate, box) {
			visit(candidate)
		}
		return true
	})
}

// PlayerAttackCircleOverlaps5180B0 applies the original shape filter once.
func PlayerAttackCircleOverlaps5180B0(center *types.Pointf, radius float32, candidate *Object, box func(types.Pointf, float32, *Object) (float64, types.Pointf)) bool {
	if center == nil || candidate == nil {
		return false
	}
	// C0|C3 rejects negative, equal and unordered penetrations.
	if candidate.Shape.Kind == ShapeKindBox {
		penetration, _ := box(*center, radius, candidate)
		return penetration > 0
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	// X is spilled to binary32. Y is multiplied by its binary32 copy
	// while it remains on the x87 stack. FSQRT also stays on the stack.
	dx := float64(spill(add(float64(center.X), -float64(candidate.PosVec.X))))
	dy := add(float64(center.Y), -float64(candidate.PosVec.Y))
	sy := float64(spill(dy))
	distance := monsterMoveForceSqrtChop53_50D50C(add(mul(dy, sy), mul(dx, dx)))
	if candidate.Shape.Kind == ShapeKindCircle {
		distance = add(distance, -float64(candidate.Shape.Circle.R))
	}
	return add(float64(radius), -distance) > 0
}
