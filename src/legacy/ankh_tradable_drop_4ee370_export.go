package legacy

/*
#include "ankh_tradable_drop_4ee370.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_dropAnkhTradable_4EE370
func nox_xxx_dropAnkhTradable_4EE370(
	owner, item *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(ankhTradableDropCall4EE370(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
