package legacy

/*
#include "barrel_collide_4eaaa0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export sub_4EAAA0
func sub_4EAAA0(source, target *C.nox_object_t, collision *C.float) {
	GetServer().S().BarrelCollide4EAAA0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
