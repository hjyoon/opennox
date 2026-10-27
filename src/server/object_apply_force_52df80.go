package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// Memory operands retain a signaling NaN until the arithmetic instruction;
// an ordinary float32-to-float64 conversion would quiet it prematurely.
func objectForceMemory32_52DF80(value float32) float64 {
	bits := math.Float32bits(value)
	if bits&0x7f800000 == 0x7f800000 && bits&0x007fffff != 0 {
		return math.Float64frombits(uint64(bits&0x80000000)<<32 | 0x7ff0000000000000 | uint64(bits&0x007fffff)<<29)
	}
	return float64(value)
}

func objectForceResult53_52DF80(a, b, result float64) float64 {
	if !math.IsNaN(result) {
		return result
	}
	const quiet = uint64(1) << 51
	const fraction = uint64(1)<<52 - 1
	x, y := math.Float64bits(a), math.Float64bits(b)
	if math.IsNaN(a) && math.IsNaN(b) {
		switch {
		case x&quiet != y&quiet:
			if x&quiet == 0 {
				x = y
			}
		case x&fraction < y&fraction:
			x = y
		case x&fraction == y&fraction:
			x &= y | ^(uint64(1) << 63)
		}
	} else if math.IsNaN(b) {
		x = y
	} else if !math.IsNaN(a) {
		// x87 invalid operations produce the negative floating-point indefinite.
		x = 0xfff8000000000000
	}
	return math.Float64frombits(x | quiet)
}

func objectForceAdd53_52DF80(a, b float64) float64 {
	return objectForceResult53_52DF80(a, b, experienceAdd53TowardZero(a, b))
}

func objectForceSub53_52DF80(a, b float64) float64 {
	return objectForceResult53_52DF80(a, b, experienceAdd53TowardZero(a, -b))
}

func objectForceMul53_52DF80(a, b float64) float64 {
	return objectForceResult53_52DF80(a, b, experienceMul53TowardZero(a, b))
}

func objectForceDiv53_52DF80(a, b float64) float64 {
	return objectForceResult53_52DF80(a, b, monsterMoveForceDivChop53_50D581(a, b))
}

// ApplyObjectForce52DF80 preserves GAME.EXE's common radial force arithmetic
// and movable gate. The caller retains the original non-missile wakeup.
// All stores and retained operations use startup PC53/ToZero locally.
func ApplyObjectForce52DF80(unit *Object, origin types.Pointf, force float32) bool {
	if !unit.IsMovable() {
		return false
	}
	dxWide := objectForceSub53_52DF80(float64(unit.PosVec.X), objectForceMemory32_52DF80(origin.X))
	dx := experienceSpill32TowardZero(dxWide)
	dyWide := objectForceSub53_52DF80(float64(unit.PosVec.Y), objectForceMemory32_52DF80(origin.Y))
	dy := experienceSpill32TowardZero(dyWide)
	// FST at 0052DFA7 spills Y without popping the unrounded difference.
	squareY := objectForceMul53_52DF80(dyWide, float64(dy))
	squareX := objectForceMul53_52DF80(float64(dx), float64(dx))
	square := objectForceAdd53_52DF80(squareY, squareX)
	distance := experienceSpill32TowardZero(objectForceAdd53_52DF80(
		monsterMoveForceSqrtChop53_50D50C(square), float64(float32(0.1))))
	scaled := experienceSpill32TowardZero(objectForceMul53_52DF80(float64(force), 10))
	factor := objectForceDiv53_52DF80(float64(scaled), float64(unit.Mass))
	x := objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(dx), factor), float64(distance))
	unit.ForceVec.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(unit.ForceVec.X)))
	y := objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(dy), factor), float64(distance))
	unit.ForceVec.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(unit.ForceVec.Y)))
	return true
}
