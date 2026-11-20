package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

var mapHazardTypeByName518770 = func(name string) int32 {
	return int32(GetServer().S().Types.IndByID(name))
}

//export sub_518770
func sub_518770() C.int {
	names := [...]string{"TeleportPentagram", "PressurePlate", "Spike", "PeriodicSpike"}
	for i, name := range names {
		id := mapHazardTypeByName518770(name)
		*memmap.PtrUint32(0x5D4594, 2386972+4*uintptr(i)) = uint32(id)
		if id == 0 {
			return 0
		}
	}
	return 1
}

//export nox_xxx_getQuestStage_51A930
func nox_xxx_getQuestStage_51A930() C.int {
	return C.int(memmap.Uint32(0x5D4594, 2388660))
}

//export sub_51A940
func sub_51A940(value C.int) C.int {
	*memmap.PtrUint32(0x5D4594, 2388656) = uint32(value)
	return value
}

//export sub_51A950
func sub_51A950() C.int {
	return C.int(memmap.Uint32(0x5D4594, 2388656))
}
