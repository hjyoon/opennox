package legacy

/*
#include "barrel_collide_4eaaa0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var barrelCollideCall4EAAA0 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().BarrelCollide4EAAA0(source, target, (*types.Pointf)(collision))
}

//export sub_4EAAA0
func sub_4EAAA0(source, target *C.nox_object_t, collision *C.float) {
	barrelCollideCall4EAAA0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
