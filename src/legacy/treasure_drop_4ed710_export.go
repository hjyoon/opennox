package legacy

/*
#include "treasure_drop_4ed710.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_dropTreasure_4ED710
func nox_xxx_dropTreasure_4ED710(
	owner, treasure *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(treasureDropCall4ED710(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(treasure)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
