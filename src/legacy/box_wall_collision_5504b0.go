package legacy

/*
#include "defs.h"
#include "GAME5.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export sub_5504B0_go
func sub_5504B0_go(obj *nox_object_t) {
	GetServer().S().BoxWallCollision5504B0(asObjectS(obj),
		func(first *server.Object, impulse types.Pointf) {
			C.nox_xxx_collSysAddCollision_548630(asObjectC(first), 0, (*C.float2)(unsafe.Pointer(&impulse)))
		},
		func(grid image.Point, collided *server.Object) {
			point := C.int2{field_0: C.int(grid.X), field_4: C.int(grid.Y)}
			C.sub_548100(&point, asObjectC(collided))
		},
	)
}
