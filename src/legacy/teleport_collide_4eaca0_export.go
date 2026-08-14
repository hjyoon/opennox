package legacy

/*
#include "teleport_collide_4eaca0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export sub_4EACA0
func sub_4EACA0(source, target *C.nox_object_t, collision *C.float) {
	srv := GetServer()
	srv.S().TeleportCollide4EACA0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.TeleportCollideRuntime4EACA0{
			Teleport: func(obj *server.Object, destination *types.Pointf) {
				teleportToMBObject4E7190(obj, func(got *server.Object) {
					Nox_xxx_unitMove_4E7010(got, *destination)
				})
			},
		},
	)
}
