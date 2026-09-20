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

var sparkCollideCall4EA300 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().SparkCollide4EA300(
		source,
		target,
		(*types.Pointf)(collision),
		server.SparkCollideRuntime4EA300{
			WallReflect: wallReflectCollideRuntime4E9D80(srv),
		},
	)
}

//export nox_xxx_collideSpark_4EA300
func nox_xxx_collideSpark_4EA300(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	sparkCollideCall4EA300(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
