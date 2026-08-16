package legacy

/*
#include "object_force_drop_4ed930.h"
*/
import "C"

//export nox_xxx_invForceDropItem_4ED930
func nox_xxx_invForceDropItem_4ED930(
	owner, item *C.nox_object_t,
) C.int {
	return C.int(objectForceDropCall4ED930(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
