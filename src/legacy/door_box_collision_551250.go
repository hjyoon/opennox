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

//export sub_551250_go
func sub_551250_go(door, box *nox_object_t, moveBox C.int) {
	GetServer().S().DoorBoxCollision551250(asObjectS(door), asObjectS(box), moveBox != 0,
		func(first, second *server.Object, normal types.Pointf) {
			C.nox_xxx_collSysAddCollision_548630(asObjectC(first), C.uintptr_t(uintptr(second.CObj())), (*C.float2)(unsafe.Pointer(&normal)))
		},
		func(obj *server.Object) {
			C.nox_xxx_unitHasCollideOrUpdateFn_537610(asObjectC(obj))
		},
		func(update *server.DoorUpdateData) {
			C.sub_548830(C.uintptr_t(uintptr(unsafe.Pointer(update))))
		},
	)
}
