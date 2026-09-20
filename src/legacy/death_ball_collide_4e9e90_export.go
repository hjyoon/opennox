package legacy

/*
#include "death_ball_collide_4e9e90.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var deathBallCollideCall4E9E90 = func(source, target *server.Object, collision unsafe.Pointer) {
	Nox_xxx_collideDeathBall_4E9E90(source, target, (*types.Pointf)(collision))
}

//export nox_xxx_collideDeathBall_4E9E90
func nox_xxx_collideDeathBall_4E9E90(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	deathBallCollideCall4E9E90(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
