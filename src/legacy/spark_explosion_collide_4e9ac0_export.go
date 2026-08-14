package legacy

/*
#include "spark_explosion_collide_4e9ac0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_fireballCollide_4E9AC0
func nox_xxx_fireballCollide_4E9AC0(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().SparkExplosionCollide4E9AC0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.SparkExplosionCollideRuntime4E9AC0{
			CheckDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
				return twoPointsAndDirection4E6E50(first, int32(direction), second)
			},
			MapPushUnits: func(
				pos types.Pointf,
				first, second, force float32,
				source *server.Object,
				arg6, arg7 int32,
			) {
				Nox_xxx_mapPushUnitsAround_52E040(pos, first, second, force, source, int(arg6), int(arg7))
			},
			MapDamageUnits: func(
				pos types.Pointf,
				radius, inner float32,
				damage int32,
				damageType object.DamageType,
				source, excluded *server.Object,
			) {
				srv.Nox_xxx_mapDamageUnitsAround(
					pos,
					radius,
					inner,
					int(damage),
					damageType,
					source,
					excluded,
					GetDoDamageWalls(),
				)
			},
			Scorch: func(pos types.Pointf, kind int32) {
				Nox_xxx_sMakeScorch_537AF0(pos, int(kind))
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
