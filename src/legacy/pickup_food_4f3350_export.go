package legacy

/*
#include "pickup_food_4f3350.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupFoodCall4F3350(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupFood_4F3350(owner, item, arg3, arg4)
}

func pickupFoodExportCall4F3350(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupFood_4F3350(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_pickupFood_4F3350
func nox_xxx_pickupFood_4F3350(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupFoodCall4F3350(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
