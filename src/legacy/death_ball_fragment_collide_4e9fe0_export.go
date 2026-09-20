package legacy

/*
#include "death_ball_fragment_collide_4e9fe0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var deathBallFragmentCollideCall4E9FE0 = func(source, target *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().DeathBallFragmentCollide4E9FE0(
		source,
		target,
		(*types.Pointf)(collision),
		server.DeathBallFragmentCollideRuntime4E9FE0{
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}

//export nox_xxx_collideDeathBallFragment_4E9FE0
func nox_xxx_collideDeathBallFragment_4E9FE0(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	deathBallFragmentCollideCall4E9FE0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
