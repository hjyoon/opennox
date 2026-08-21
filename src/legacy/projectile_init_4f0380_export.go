package legacy

/*
#include "projectile_init_4f0380.h"
*/
import "C"

//export nox_xxx_unitProjectileInit_4F0380
func nox_xxx_unitProjectileInit_4F0380(unit *C.nox_object_t) {
	projectileInitCall4F0380(asObjectS((*nox_object_t)(unit)))
}
