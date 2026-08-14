package legacy

/*
#include "webbing_collide_4ea380.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideWebbing_4EA380
func nox_xxx_collideWebbing_4EA380(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().WebbingCollide4EA380(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.WebbingCollideRuntime4EA380{
			DelayedDelete: srv.DelayedDelete,
			ApplyEnchant: func(obj *server.Object, enchant server.EnchantID, duration, power uint32) {
				Nox_xxx_buffApplyTo_4FF380(obj, enchant, int(uint16(duration)), int(uint8(power)))
			},
		},
	)
}
