package opennox

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

// These operations model GAME.EXE's gameplay x87 environment: CRT startup
// 004031F2 selects 53-bit precision and each main-loop iteration at 0043E2C1
// selects round-toward-zero. The location dependencies retain every arithmetic
// result until FCOMP, without a binary32 spill. Float32 coordinates bound the
// differences, squares and exact residuals well inside binary64's exponent
// range. Correct nearest-even operations locally; do not change the thread's
// floating-point environment or reuse a nearest-even distance helper.

//go:noinline
func aiDependencyAddChop53_546B63(a, b float64) float64 {
	sum := float64(a + b)
	// TwoSum recovers the exact rounding error, also for operands separated
	// by more than 53 bits. Nonfinite inputs produce an unordered residual
	// and keep their original infinity/NaN result.
	bVirtual := float64(sum - a)
	error := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && error < 0 || sum < 0 && error > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func aiDependencySquareChop53_546B63(value float64) float64 {
	square := float64(value * value)
	// FMA is used only to obtain the exact multiplication residual, never
	// to contract the original separate square and addition instructions.
	if math.FMA(value, value, -square) < 0 {
		return math.Nextafter(square, 0)
	}
	return square
}

//go:noinline
func aiDependencySqrtChop53_546B63(square float64) float64 {
	root := math.Sqrt(square)
	if math.FMA(root, root, -square) > 0 {
		return math.Nextafter(root, 0)
	}
	return root
}

// aiDependencyLocation546B63 preserves conditions 51 and 52 of 00546A70.
// Read the entry-cached slot's X before unit X, then slot Y before unit Y;
// only read radius after the retained Y-square, X-square, sum and FSQRT.
// Direct PosVec reads keep missing-unit faults instead of using nil-safe Pos.
func aiDependencyLocation546B63(unit *server.Object, slot *server.AIStackItem, closer bool) bool {
	dx := aiDependencyAddChop53_546B63(float64(slot.ArgF32(2)), -float64(unit.PosVec.X))
	dy := aiDependencyAddChop53_546B63(float64(slot.ArgF32(3)), -float64(unit.PosVec.Y))
	ySquare := aiDependencySquareChop53_546B63(dy)
	xSquare := aiDependencySquareChop53_546B63(dx)
	distance := aiDependencySqrtChop53_546B63(aiDependencyAddChop53_546B63(ySquare, xSquare))
	radius := float64(slot.ArgF32(0))
	if closer {
		// FCOMP 00546BAE tests C0|C3: less, equal and unordered all pass.
		return !(distance > radius)
	}
	// FCOMP 00546B7F tests C0 only: only ordered >= passes.
	return distance >= radius
}
