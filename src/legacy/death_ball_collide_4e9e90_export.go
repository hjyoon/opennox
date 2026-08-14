package legacy

/*
#include "death_ball_collide_4e9e90.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_collideDeathBall_4E9E90
func nox_xxx_collideDeathBall_4E9E90(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	Nox_xxx_collideDeathBall_4E9E90(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
