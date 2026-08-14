package legacy

/*
#include "bear_trap_collide_4eb890.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideBearTrap_4EB890
func nox_xxx_collideBearTrap_4EB890(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().BearTrapCollide4EB890(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.BearTrapCollideRuntime4EB890{
			CreateAt: func(item, owner *server.Object, pos types.Pointf) {
				srv.CreateObjectAt(item, owner, pos)
			},
			DelayedDelete: srv.DelayedDelete,
			ApplyEnchant: func(obj *server.Object, enchant server.EnchantID, duration, power uint32) {
				Nox_xxx_buffApplyTo_4FF380(obj, enchant, int(uint16(duration)), int(uint8(power)))
			},
		},
	)
}
