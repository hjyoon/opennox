package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export sub_5488B0_go
func sub_5488B0_go(door, circle *nox_object_t, moveCircle C.int) {
	GetServer().S().DoorCircleCollision5488B0(asObjectS(door), asObjectS(circle), moveCircle != 0,
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
