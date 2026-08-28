package legacy

/*
#include "pickup_potion_4f37d0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupPotionCall4F37D0(
	owner, potion *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupPotion_4F37D0(owner, potion, arg3, arg4)
}

func pickupPotionExportCall4F37D0(
	owner, potion *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupPotion_4F37D0(
		asObjectC(owner),
		asObjectC(potion),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_pickupPotion_4F37D0
func nox_xxx_pickupPotion_4F37D0(
	owner, potion *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupPotionCall4F37D0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(potion)),
		int32(arg3),
		int32(arg4),
	))
}
