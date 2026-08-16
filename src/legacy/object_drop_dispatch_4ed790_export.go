package legacy

/*
#include "object_drop_dispatch_4ed790.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_drop_4ED790
func nox_xxx_drop_4ED790(
	owner, item *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(objectDropDispatchCall4ED790(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
