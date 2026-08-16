package legacy

/*
#include "crown_drop_4ed5e0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_dropCrown_4ED5E0
func nox_xxx_dropCrown_4ED5E0(
	owner, crown *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(crownDropCall4ED5E0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(crown)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
