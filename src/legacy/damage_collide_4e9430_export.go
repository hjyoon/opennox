package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var damageCollideCall4E9430 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().DamageCollide4E9430(source, target, collision)
}

//export nox_xxx_collideDamage_4E9430
func nox_xxx_collideDamage_4E9430(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	damageCollideCall4E9430(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
