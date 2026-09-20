package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

var projectileSparkCollideCall4E8880 = func(projectile, other *server.Object, collision unsafe.Pointer) {
	srv := GetServer()
	srv.S().ProjectileSparkCollide4E8880(
		projectile,
		other,
		collision,
		server.ProjectileSparkCollideRuntime4E8880{
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}

//export nox_xxx_collideProjectileSpark_4E8880
func nox_xxx_collideProjectileSpark_4E8880(
	projectile, other *C.nox_object_t,
	collision *C.float,
) {
	projectileSparkCollideCall4E8880(
		asObjectS((*nox_object_t)(projectile)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
	)
}
