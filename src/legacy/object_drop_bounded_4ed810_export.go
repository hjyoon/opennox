package legacy

/*
#include "object_drop_bounded_4ed810.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_drop_4ED810
func nox_xxx_drop_4ED810(
	owner, item *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(objectDropBoundedCall4ED810(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
