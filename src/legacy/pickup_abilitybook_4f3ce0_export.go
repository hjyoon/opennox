package legacy

/*
#include "pickup_abilitybook_4f3ce0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupAbilityBookCall4F3CE0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupAbilitybook_4F3CE0(owner, item, arg3, arg4)
}

func pickupAbilityBookExportCall4F3CE0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupAbilitybook_4F3CE0(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_pickupAbilitybook_4F3CE0
func nox_xxx_pickupAbilitybook_4F3CE0(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupAbilityBookCall4F3CE0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
