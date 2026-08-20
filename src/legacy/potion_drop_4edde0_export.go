package legacy

/*
#include "potion_drop_4edde0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export sub_4EDDE0
func sub_4EDDE0(
	owner, potion *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(potionDropCall4EDDE0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(potion)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
