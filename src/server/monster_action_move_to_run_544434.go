package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// GAME.EXE's gameplay x87 environment uses precision 53 (CRT 004031F2)
// and round-toward-zero (main loop 0043E2C1). The float32 source coordinates
// bound all exact residuals well inside binary64's exponent range. Correct
// nearest-even arithmetic locally, without changing the thread environment.
//
//go:noinline
func monsterMoveToRunAddChop53_544434(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	error := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && error < 0 || sum < 0 && error > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func monsterMoveToRunSquareChop53_544434(value float64) float64 {
	square := float64(value * value)
	// FMA recovers only the exact residual, not a contracted distance sum.
	if math.FMA(value, value, -square) < 0 {
		return math.Nextafter(square, 0)
	}
	return square
}

// monsterMoveToRunSpill544440 models the single FST float32 near-threshold
// spill. Finite overflow under chop stores max-finite, not infinity. The
// unspilled register remains live for the separate far threshold.
func monsterMoveToRunSpill544440(value float64) float64 {
	narrow := float32(value)
	if math.Abs(float64(narrow)) > math.Abs(value) {
		narrow = math.Nextafter32(narrow, 0)
	}
	return float64(narrow)
}

// monsterMoveToRunBand544434 returns -1 (stop), 0 (retain), or 1 (start)
// for the ESCORT prefix of 005443F0. Constants 00583D04/00583C08 are the
// original float32 words 40400000 (3) and 41F00000 (30). No FSQRT occurs.
func monsterMoveToRunBand544434(unit, target types.Pointf, followRange float32) int {
	// A float32 times integer 3 is exact at 53-bit precision.
	product := float64(followRange) * 3
	near := monsterMoveToRunSpill544440(product)
	far := monsterMoveToRunAddChop53_544434(product, 30)
	dx := monsterMoveToRunAddChop53_544434(float64(target.X), -float64(unit.X))
	dy := monsterMoveToRunAddChop53_544434(float64(target.Y), -float64(unit.Y))
	ySquare := monsterMoveToRunSquareChop53_544434(dy)
	xSquare := monsterMoveToRunSquareChop53_544434(dx)
	square := monsterMoveToRunAddChop53_544434(ySquare, xSquare)
	// FCOMPP 0054446E tests C0 only: less and unordered stop running.
	if !(square >= monsterMoveToRunSquareChop53_544434(near)) {
		return -1
	}
	// FCOMPP 00544489 tests C0|C3: only ordered greater starts running.
	if square > monsterMoveToRunSquareChop53_544434(far) {
		return 1
	}
	return 0
}
