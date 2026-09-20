package legacy

/*
#include "wall_reflect_collide_4e9d80.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func wallReflectCollideRuntime4E9D80(srv Server) server.WallReflectCollideRuntime4E9D80 {
	return server.WallReflectCollideRuntime4E9D80{
		YellowStarCollide: C.nox_xxx_collideSulphurShot_4E9E50,
		DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
			srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
		},
		DelayedDelete: srv.DelayedDelete,
	}
}

var wallReflectCollideCall4E9D80 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().WallReflectCollide4E9D80(
		source,
		target,
		(*types.Pointf)(collision),
		wallReflectCollideRuntime4E9D80(srv),
	)
}

var yellowStarShotCollideCall4E9E50 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().YellowStarShotCollide4E9E50(
		source,
		target,
		(*types.Pointf)(collision),
		wallReflectCollideRuntime4E9D80(srv),
	)
}

//export nox_xxx_collideSulphurShot2_4E9D80
func nox_xxx_collideSulphurShot2_4E9D80(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	wallReflectCollideCall4E9D80(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}

//export nox_xxx_collideSulphurShot_4E9E50
func nox_xxx_collideSulphurShot_4E9E50(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	yellowStarShotCollideCall4E9E50(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
