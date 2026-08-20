package legacy

/*
#include "server__gamemech__explevel.h"
*/
import "C"

//export nox_xxx_plyrGiveExp_4EF3A0_exp_level
func nox_xxx_plyrGiveExp_4EF3A0_exp_level(unit *C.nox_object_t, award C.float) {
	directExperienceGrantCall4EF3A0(
		asObjectS((*nox_object_t)(unit)),
		float32(award),
	)
}
