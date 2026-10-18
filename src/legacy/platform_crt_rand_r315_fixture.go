package legacy

/*
#include "GAME4_2.h"
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
