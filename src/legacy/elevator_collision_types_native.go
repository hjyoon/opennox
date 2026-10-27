package legacy

/*
#include "GAME5.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

// Initialization runs after the collision caller caches its update pointer.
func elevatorCollisionTypeNative551BF0(name string) int {
	if memmap.Uint32(0x5D4594, 2491808) == 0 {
		C.sub_551BF0()
	}
	var index uintptr
	switch name {
	case "SmallFist":
		index = 0
	case "MediumFist":
		index = 1
	case "LargeFist":
		index = 2
	case "Meteor":
		index = 3
	default:
		panic("unknown elevator collision type")
	}
	return int(memmap.Uint32(0x5D4594, 2491792+index*4))
}
