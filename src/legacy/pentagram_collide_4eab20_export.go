package legacy

/*
#include "pentagram_collide_4eab20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var pentagramCollideCall4EAB20 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().PentagramCollide4EAB20(source, target, (*types.Pointf)(collision))
}

//export nox_xxx_collidePentagram_4EAB20
func nox_xxx_collidePentagram_4EAB20(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	pentagramCollideCall4EAB20(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
