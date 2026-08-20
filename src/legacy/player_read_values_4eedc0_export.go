package legacy

/*
#include "player_read_values_4eedc0.h"
*/
import "C"

//export nox_xxx_plrReadVals_4EEDC0
func nox_xxx_plrReadVals_4EEDC0(
	unit *C.nox_object_t,
	rewardArg C.int32_t,
) C.int32_t {
	return C.int32_t(playerReadValuesCall4EEDC0(
		asObjectS((*nox_object_t)(unit)),
		int32(rewardArg),
	))
}
