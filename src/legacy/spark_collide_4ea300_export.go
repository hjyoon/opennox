package legacy

/*
#include "spark_collide_4ea300.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideSpark_4EA300
func nox_xxx_collideSpark_4EA300(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().SparkCollide4EA300(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.SparkCollideRuntime4EA300{
			WallReflect: wallReflectCollideRuntime4E9D80(srv),
		},
	)
}
