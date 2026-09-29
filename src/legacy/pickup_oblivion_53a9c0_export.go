package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupOblivionCall53A9C0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupOblivion_53A9C0(owner, item, arg3, arg4)
}

func pickupOblivionExportCall53A9C0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_sendMsgOblivionPickup_53A9C0(
		asObjectC(owner),
		asObjectC(item),
		C.int(arg3),
		C.int(arg4),
	))
}

//export nox_xxx_sendMsgOblivionPickup_53A9C0_go
func nox_xxx_sendMsgOblivionPickup_53A9C0_go(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int,
) C.int {
	return C.int(pickupOblivionCall53A9C0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
