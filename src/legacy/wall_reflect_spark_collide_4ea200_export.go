package legacy

/*
#include "wall_reflect_spark_collide_4ea200.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideWallReflectSpark_4EA200
func nox_xxx_collideWallReflectSpark_4EA200(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().WallReflectSparkCollide4EA200(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.WallReflectSparkCollideRuntime4EA200{
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
