package legacy

/*
#include "own_collide_4ea2c0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var ownCollideCall4EA2C0 = func(source, target *server.Object, _ unsafe.Pointer) {
	GetServer().S().OwnCollide4EA2C0(source, target)
}

//export sub_4EA2C0
func sub_4EA2C0(source, target *C.nox_object_t, collision *C.float) {
	ownCollideCall4EA2C0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
