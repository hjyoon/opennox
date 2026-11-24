package legacy

/*
#include <float.h>
#include "GAME4_2.h"

enum { nox_mapgen_random_float_evaluation = FLT_EVAL_METHOD };
*/
import "C"

import (
	"math"
	"math/big"
	"runtime"
)

const (
	mapgenFloatRandomScale = 0.000030518509
	mapgenFloatEvalMethod  = C.nox_mapgen_random_float_evaluation
)

func mapgenExtendedFloatRandom(draw uint32, low, high float64) float64 {
	// A 64-bit significand preserves the intermediate x87 operations.
	var minimum, difference, factor big.Float
	minimum.SetPrec(64).SetFloat64(low)
	difference.SetPrec(64).SetFloat64(high)
	factor.SetPrec(64).SetUint64(uint64(draw))
	difference.Sub(&difference, &minimum)
	difference.Mul(&difference, &factor)
	factor.SetFloat64(mapgenFloatRandomScale)
	difference.Mul(&difference, &factor)
	difference.Add(&difference, &minimum)
	result, _ := difference.Float64()
	return result
}

func mapgenFloatRandomFromDraw(draw uint32, low, high float32) float64 {
	minimum, maximum := float64(low), float64(high)
	width := float64(float32(high - low))
	if mapgenFloatEvalMethod > 0 {
		width = maximum - minimum
	}
	if mapgenFloatEvalMethod == 2 && !math.IsNaN(width) && !math.IsInf(width, 0) {
		return mapgenExtendedFloatRandom(draw, minimum, maximum)
	}
	scaled := float64(float64(draw) * width)
	if runtime.GOARCH == "arm64" {
		return math.FMA(scaled, mapgenFloatRandomScale, minimum)
	}
	return float64(scaled*mapgenFloatRandomScale) + minimum
}

//export sub_526BC0
func sub_526BC0(low, high C.float) C.double {
	return C.double(mapgenFloatRandomFromDraw(uint32(nox_platform_rand()), float32(low), float32(high)))
}
