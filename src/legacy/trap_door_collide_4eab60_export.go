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

//export nox_xxx_collideTrapDoor_4EAB60
func nox_xxx_collideTrapDoor_4EAB60(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().TrapDoorCollide4EAB60(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.TrapDoorCollideRuntime4EAB60{
			ScriptCallback: srv.NoxScriptC().ScriptCallback,
		},
	)
}
