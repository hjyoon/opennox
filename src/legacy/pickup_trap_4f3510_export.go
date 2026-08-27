package legacy

/*
#include "pickup_trap_4f3510.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupTrapCall4F3510(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupTrap_4F3510(owner, item, arg3, arg4)
}

func pickupTrapExportCall4F3510(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupTrap_4F3510(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_pickupTrap_4F3510
func nox_xxx_pickupTrap_4F3510(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupTrapCall4F3510(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
