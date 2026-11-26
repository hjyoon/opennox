package legacy

/*
#include "GAME3_3.h"
#include "GAME5_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var manaDrainCollideCall4E9490 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().ManaDrainCollide4E9490(
		source,
		target,
		collision,
		func(token uint32, delta int16) {
			addPlayerProtectedStat(token, uint32(delta))
		},
	)
}

//export nox_xxx_collideManadrain_4E9490
func nox_xxx_collideManadrain_4E9490(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	manaDrainCollideCall4E9490(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
