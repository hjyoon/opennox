package legacy

/*
#include "pentagram_collide_4eab20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_collidePentagram_4EAB20
func nox_xxx_collidePentagram_4EAB20(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().PentagramCollide4EAB20(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
