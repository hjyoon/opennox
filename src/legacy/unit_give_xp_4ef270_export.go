package legacy

/*
#include "server__object__health.h"
*/
import "C"

//export nox_xxx_unitGiveXP_4EF270
func nox_xxx_unitGiveXP_4EF270(unit *C.nox_object_t, target C.float) C.double {
	return C.double(unitGiveXPCall4EF270(
		asObjectS((*nox_object_t)(unit)),
		float32(target),
	))
}
