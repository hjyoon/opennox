package legacy

/*
#include "GAME4_2.h"
*/
import "C"

//export nox_xxx_mapGenSetRngSeed_526AB0
func nox_xxx_mapGenSetRngSeed_526AB0(seed C.uint) {
	nox_platform_srand(seed)
}

//export nox_xxx_mapGenRandFunc_526AC0
func nox_xxx_mapGenRandFunc_526AC0(low, high C.int) C.int {
	span := int32(high - low + 1)
	// The original converts the wrapped product to unsigned for division.
	offset := uint32(span*nox_platform_rand()) / 0x7fff
	return min(C.int(uint32(low)+offset), high)
}

//export nox_xxx_mapGenRandFunc2_526B00
func nox_xxx_mapGenRandFunc2_526B00(center, spread C.int) C.int {
	if spread == 0 {
		return center
	}
	limit := 10000 / spread
	for {
		value := nox_xxx_mapGenRandFunc_526AC0(-spread, spread)
		weight := nox_xxx_mapGenRandFunc_526AC0(0, limit)
		if weight < limit*(spread-value)/spread && weight < limit*(value+spread)/spread {
			return value + center
		}
	}
}
