package legacy

/*
#include "server__gamemech__explevel.h"
*/
import "C"

//export sub_4EF2E0_exp_level
func sub_4EF2E0_exp_level(unit *C.nox_object_t) {
	experienceLevelUpdateCall4EF2E0(
		asObjectS((*nox_object_t)(unit)),
	)
}
