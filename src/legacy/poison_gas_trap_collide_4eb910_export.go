package legacy

/*
#include "poison_gas_trap_collide_4eb910.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collidePoisonGasTrap_4EB910
func nox_xxx_collidePoisonGasTrap_4EB910(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().PoisonGasTrapCollide4EB910(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.PoisonGasTrapCollideRuntime4EB910{
			CreateAt: func(item, owner *server.Object, pos types.Pointf) {
				srv.CreateObjectAt(item, owner, pos)
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
