package legacy

/*
#include "fist_collide_4eadf0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var fistCollideCall4EADF0 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().FistCollide4EADF0(source, target, (*types.Pointf)(collision))
}

//export nox_xxx_collideFist_4EADF0
func nox_xxx_collideFist_4EADF0(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	fistCollideCall4EADF0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
