package legacy

/*
#include "GAME4_3.h"
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export sub_550F80_go
func sub_550F80_go(first, second *nox_object_t) {
	GetServer().S().BoxBoxCollision550F80(asObjectS(first), asObjectS(second),
		func(collisionFirst, collisionSecond *server.Object, normal types.Pointf) {
			C.nox_xxx_collSysAddCollision_548630(asObjectC(collisionFirst), C.uintptr_t(uintptr(collisionSecond.CObj())), (*C.float2)(unsafe.Pointer(&normal)))
		},
		func(obj *server.Object) {
			C.nox_xxx_unitHasCollideOrUpdateFn_537610(asObjectC(obj))
		},
	)
}
