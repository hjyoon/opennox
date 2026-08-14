package legacy

/*
#include "pixie_collide_4ea080.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collidePixie_4EA080
func nox_xxx_collidePixie_4EA080(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().PixieCollide4EA080(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.PixieCollideRuntime4EA080{
			CheckDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
				return twoPointsAndDirection4E6E50(first, int32(direction), second)
			},
			ChangeOwner: Nox_xxx_changeOwner_52BE40,
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DelayedDelete:   srv.DelayedDelete,
			InversionEffect: InversionEffectPointer4E03D0(),
		},
	)
}
