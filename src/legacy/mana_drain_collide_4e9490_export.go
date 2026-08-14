package legacy

/*
#include "GAME3_3.h"
#include "GAME5_2.h"
*/
import "C"

import "unsafe"

//export nox_xxx_collideManadrain_4E9490
func nox_xxx_collideManadrain_4E9490(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	GetServer().S().ManaDrainCollide4E9490(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
		func(token uint32, delta int16) {
			C.nox_xxx_protectMana_56F9E0(C.int(token), C.short(delta))
		},
	)
}
