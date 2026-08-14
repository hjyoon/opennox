package legacy

/*
#include "die_collide_4e99b0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideDie_4E99B0
func nox_xxx_collideDie_4E99B0(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().DieCollide4E99B0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
		server.DieCollideRuntime4E99B0{DelayedDelete: srv.DelayedDelete},
	)
}
