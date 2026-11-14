package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerCanTalkMB_57A160
func nox_xxx_playerCanTalkMB_57A160(pl *C.nox_playerInfo) C.int {
	if server.PlayerCantTalkNative57A160((*server.Player)(unsafe.Pointer(pl))) {
		return 1
	}
	return 0
}

//export nox_xxx_collideReflect_57B810
func nox_xxx_collideReflect_57B810(normal, velocity *C.float) *C.float {
	server.ReflectVelocityNative57B810((*types.Pointf)(unsafe.Pointer(normal)), (*types.Pointf)(unsafe.Pointer(velocity)))
	return velocity
}
