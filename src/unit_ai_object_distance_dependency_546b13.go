package opennox

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

// aiDependencyBoxHeightHalf546B13 models the m32real spills at 004E6C54
// and 004E6CA1 in GAME.EXE's gameplay round-toward-zero environment. The
// binary64 half is exact for every finite binary32 height; only the store
// can round it. Correct a nearest-even store that rounded away from zero.
func aiDependencyBoxHeightHalf546B13(height float32) float64 {
	half := float64(height) * 0.5
	rounded := float32(half)
	if half > 0 && float64(rounded) > half || half < 0 && float64(rounded) < half {
		rounded = math.Nextafter32(rounded, 0)
	}
	return float64(rounded)
}

func aiDependencySubtractShape546B13(distance float64, shape *server.Shape, kind server.ShapeKind) float64 {
	switch kind {
	case server.ShapeKindCircle:
		return aiDependencyAddChop53_546B63(distance, -float64(shape.Circle.R))
	case server.ShapeKindBox:
		width := float64(shape.Box.W) * 0.5
		height := aiDependencyBoxHeightHalf546B13(shape.Box.H)
		// Original C0|C3: choose width only for ordered strict >.
		extent := height
		if width > height {
			extent = width
		}
		return aiDependencyAddChop53_546B63(distance, -extent)
	default:
		return distance
	}
}

// aiDependencyObjectSurfaceDistance546B13 retains 004E6C00's x87 return
// through the dependency FCOMP, unlike the legacy binary32 API wrapper.
// Use the original 53-bit/chop operations without changing the thread FPU.
func aiDependencyObjectSurfaceDistance546B13(unit, target *server.Object) float64 {
	dx := aiDependencyAddChop53_546B63(float64(unit.PosVec.X), -float64(target.PosVec.X))
	dy := aiDependencyAddChop53_546B63(float64(unit.PosVec.Y), -float64(target.PosVec.Y))
	unitKind := unit.Shape.Kind // 004E6C14 precedes the retained arithmetic.
	ySquare := aiDependencySquareChop53_546B63(dy)
	xSquare := aiDependencySquareChop53_546B63(dx)
	distance := aiDependencySqrtChop53_546B63(aiDependencyAddChop53_546B63(ySquare, xSquare))
	distance = aiDependencySubtractShape546B13(distance, &unit.Shape, unitKind)
	distance = aiDependencySubtractShape546B13(distance, &target.Shape, target.Shape.Kind)
	minimum := float64(float32(0.01)) // Original DWORD at 00583A9C.
	// The clamp tests C0 only, so less and unordered both use minimum.
	if !(distance >= minimum) {
		return minimum
	}
	return distance
}

func aiDependencyObjectDistance546B13(unit *server.Object, slot *server.AIStackItem, closer bool) bool {
	target := slot.ArgObj(2)
	if target == nil {
		return false
	}
	// Original reads radius only after the complete collision-distance call.
	distance := aiDependencyObjectSurfaceDistance546B13(unit, target)
	radius := float64(slot.ArgF32(0))
	if closer {
		return !(distance > radius) // 00546B55: less/equal/unordered pass.
	}
	return distance > radius // 00546B2D: only ordered strict greater passes.
}
