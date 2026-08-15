package legacy

/*
#include "undead_killer_collide_4ebd40.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideUndeadKiller_4EBD40
func nox_xxx_collideUndeadKiller_4EBD40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().UndeadKillerCollide4EBD40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.UndeadKillerCollideRuntime4EBD40{DelayedDelete: srv.DelayedDelete},
	)
}
