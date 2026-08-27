package legacy

/*
#include "inventory_serv_place_4f36f0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func inventoryServPlaceCall4F36F0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_inventoryServPlaceRaw_4F36F0(owner, item, arg3, arg4)
}

func inventoryServPlaceExportCall4F36F0(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_inventoryServPlace_4F36F0(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_inventoryServPlace_4F36F0
func nox_xxx_inventoryServPlace_4F36F0(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(inventoryServPlaceCall4F36F0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
