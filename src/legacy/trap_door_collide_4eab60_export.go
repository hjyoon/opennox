package legacy

/*
#include "trap_door_collide_4eab60.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var trapDoorCollideCall4EAB60 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().TrapDoorCollide4EAB60(
		source,
		target,
		(*types.Pointf)(collision),
		server.TrapDoorCollideRuntime4EAB60{
			ScriptCallback: srv.NoxScriptC().ScriptCallback,
		},
	)
}

//export nox_xxx_collideTrapDoor_4EAB60
func nox_xxx_collideTrapDoor_4EAB60(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	trapDoorCollideCall4EAB60(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
