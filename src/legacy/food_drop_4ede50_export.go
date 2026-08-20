package legacy

/*
#include "food_drop_4ede50.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_dropFood_4EDE50
func nox_xxx_dropFood_4EDE50(
	owner, food *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(foodDropCall4EDE50(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(food)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
