package legacy

/*
#include "aud_event_pickup_4f3d50.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func audEventPickupCall4F3D50(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_objectPickupAudEvent_4F3D50(owner, item, arg3, arg4)
}

func audEventPickupExportCall4F3D50(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_objectPickupAudEvent_4F3D50(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_objectPickupAudEvent_4F3D50
func nox_objectPickupAudEvent_4F3D50(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(audEventPickupCall4F3D50(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
