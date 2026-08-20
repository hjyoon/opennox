package legacy

/*
#include "ability_grant_4eed40.h"
*/
import "C"

//export nox_xxx_abilGivePlayerAll_4EED40
func nox_xxx_abilGivePlayerAll_4EED40(
	unit *C.nox_object_t,
	count C.int8_t,
	rewardArg C.int32_t,
) {
	abilityGivePlayerAllCall4EED40(
		asObjectS((*nox_object_t)(unit)),
		int8(count),
		int32(rewardArg),
	)
}
