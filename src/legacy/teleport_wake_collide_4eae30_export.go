package legacy

/*
#include "teleport_wake_collide_4eae30.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideTeleportWake_4EAE30
func nox_xxx_collideTeleportWake_4EAE30(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().TeleportWakeCollide4EAE30(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.TeleportWakeCollideRuntime4EAE30{
			Teleport: func(obj *server.Object, destination *types.Pointf) {
				teleportToMBObject4E7190(obj, func(got *server.Object) {
					Nox_xxx_unitMove_4E7010(got, *destination)
				})
			},
		},
	)
}
