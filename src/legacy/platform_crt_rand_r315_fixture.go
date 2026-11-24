package legacy

/*
#include "GAME4_2.h"

static double nox_test_mapgen_float_random(unsigned int draw, float low, float high) {
	return (double)draw * (high - low) * 0.000030518509 + low;
}
*/
import "C"

// Keep this regression bridge on the real C mapgen entry points. Replacing
// them with a Go model would miss the CRT C ABI that caused the prefab failure.
func mapgenRandValues315(seed uint32, count int) ([]int32, []float64) {
	C.nox_xxx_mapGenSetRngSeed_526AB0(C.uint(seed))
	ints := make([]int32, count)
	positions := make([]float64, count)
	for i := range ints {
		ints[i] = int32(C.nox_xxx_mapGenRandFunc_526AC0(-500, 500))
		positions[i] = float64(C.sub_526BC0(-512, 512))
	}
	return ints, positions
}

func mapgenRandInt315(low, high int32) int32 {
	return int32(C.nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
}

func mapgenRandWeighted315(center, spread int32) int32 {
	return int32(C.nox_xxx_mapGenRandFunc2_526B00(C.int(center), C.int(spread)))
}

func mapgenFloatRandomReference(draw uint32, low, high float32) float64 {
	return float64(C.nox_test_mapgen_float_random(C.uint(draw), C.float(low), C.float(high)))
}
