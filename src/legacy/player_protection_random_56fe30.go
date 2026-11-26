package legacy

/*
#include "GAME5_2.h"

*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var (
	playerProtectionRandomWidth   = (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2516372))
	playerProtectionRandomMinimum = (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2516380))
)

func playerProtectionRandomState() *[5]float64 {
	return memmap.PtrT[[5]float64](0x5D4594, 2516388)
}

func playerProtectionRandomConstants() *[5]float64 {
	return memmap.PtrT[[5]float64](0x581450, 11344)
}

func nextPlayerProtectionRandom(state, constants *[5]float64) float64 {
	state[3], state[2], state[1] = state[2], state[1], state[0]
	value := mulFloat64TowardZero(state[0], constants[4])
	value = addFloat64TowardZero(value, mulFloat64TowardZero(state[2], constants[3]))
	// The original uses the same shifted sample for these two terms.
	value = addFloat64TowardZero(value, mulFloat64TowardZero(state[3], constants[2]))
	value = addFloat64TowardZero(value, mulFloat64TowardZero(state[3], constants[1]))
	value = addFloat64TowardZero(value, state[4])
	whole := math.Floor(value)
	state[0] = value - whole
	state[4] = mulFloat64TowardZero(whole, constants[0])
	return state[0]
}

//export nox_xxx_unkDoubleSmth_56FE30
func nox_xxx_unkDoubleSmth_56FE30() float64 {
	return nextPlayerProtectionRandom(playerProtectionRandomState(), playerProtectionRandomConstants())
}

func seedPlayerProtectionRandom(state, constants *[5]float64, seed uint32) {
	if seed == 0 {
		seed = math.MaxUint32
	}
	for i := range state {
		seed ^= seed << 13
		seed ^= seed >> 17
		seed ^= seed << 5
		state[i] = float64(seed) * constants[0]
	}
	for range 19 {
		nextPlayerProtectionRandom(state, constants)
	}
}

//export sub_56FF00
func sub_56FF00(seed int32) {
	seedPlayerProtectionRandom(playerProtectionRandomState(), playerProtectionRandomConstants(), uint32(seed))
	*playerProtectionRandomMinimum = 0
	*memmap.PtrUint32(0x5D4594, 2516376) = 99
	*playerProtectionRandomWidth = 100
}

//export sub_56FF80
func sub_56FF80(low, high int32) int32 {
	*memmap.PtrUint32(0x5D4594, 2516376) = uint32(high)
	*playerProtectionRandomMinimum = uint32(low)
	width := uint32(high) - uint32(low) + 1
	*playerProtectionRandomWidth = width
	fraction := nextPlayerProtectionRandom(playerProtectionRandomState(), playerProtectionRandomConstants())
	value := uint32(x87TruncQwordLow(mulFloat64TowardZero(fraction, float64(width))))
	return int32(uint32(low) + min(value, width))
}
