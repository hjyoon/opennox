package legacy

/*
#include "fist_collide_4eadf0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_collideFist_4EADF0
func nox_xxx_collideFist_4EADF0(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().FistCollide4EADF0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
