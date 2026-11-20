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
	mul := pointOnLineMulChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	r := float64(radius)
	rect := types.Rectf{
		Min: types.Ptf(spill(add(float64(center.X), -r)), spill(add(float64(center.Y), -r))),
		Max: types.Ptf(spill(add(r, float64(center.X))), spill(add(r, float64(center.Y)))),
	}
	eachRect(rect, func(candidate *Object) bool {
		if candidate == nil {
			return true
		}
		var penetration float64
		if candidate.Shape.Kind == ShapeKindBox {
			// The original compares the retained box return directly to zero.
			penetration, _ = box(*center, radius, candidate)
		} else {
			// X is FSTP binary32; Y is FST binary32 but remains retained
			// for multiplication by the stored Y. FSQRT stays retained.
			dx := float64(spill(add(float64(center.X), -float64(candidate.PosVec.X))))
			dy := add(float64(center.Y), -float64(candidate.PosVec.Y))
			sy := float64(spill(dy))
			distance := monsterMoveForceSqrtChop53_50D50C(add(mul(dy, sy), mul(dx, dx)))
			if candidate.Shape.Kind == ShapeKindCircle {
				distance = add(distance, -float64(candidate.Shape.Circle.R))
			}
			penetration = add(r, -distance)
		}
		// C0|C3 rejects negative, equal and unordered penetrations.
		if penetration > 0 {
			visit(candidate)
		}
		return true
	})
}
