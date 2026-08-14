package legacy

/*
#include "sign_collide_4eab40.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_collideSign_4EAB40
func nox_xxx_collideSign_4EAB40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().SignCollide4EAB40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
