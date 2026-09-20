package legacy

/*
#include "sign_collide_4eab40.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var signCollideCall4EAB40 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().SignCollide4EAB40(source, target, (*types.Pointf)(collision))
}

//export nox_xxx_collideSign_4EAB40
func nox_xxx_collideSign_4EAB40(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	signCollideCall4EAB40(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
