package legacy

/*
#include "default_drop_4ed290.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_dropDefault_4ED290
func nox_xxx_dropDefault_4ED290(
	owner, item *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(defaultDropCall4ED290(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
