package legacy

/*
#include "own_collide_4ea2c0.h"
*/
import "C"

//export sub_4EA2C0
func sub_4EA2C0(source, target *C.nox_object_t, collision *C.float) {
	_ = collision
	GetServer().S().OwnCollide4EA2C0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
	)
}
