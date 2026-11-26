package legacy

/*
#include <stdint.h>
*/
import "C"

import "unsafe"

// Collision coefficients are extracted C-owned variables. Their retired blob
// slots do not own the live values used by the original collision system.
func collisionPushScalePtr550F80() *float32 {
	return (*float32)(unsafe.Pointer(&legacyAddressableWords.dword_587000_292488))
}

func collisionSpringPtr5488B0() *float32 {
	return (*float32)(unsafe.Pointer(&legacyAddressableWords.dword_587000_292492))
}

func collisionPushScaleNative550F80() float32 { return *collisionPushScalePtr550F80() }
func collisionSpringNative5488B0() float32    { return *collisionSpringPtr5488B0() }
