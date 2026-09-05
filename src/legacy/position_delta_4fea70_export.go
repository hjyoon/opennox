package legacy

/*
#include "position_delta_4fea70.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func positionDeltaLegacy4FEA70(object *server.Object, point *types.Pointf) int32 {
	return GetServer().S().PositionDelta4FEA70(object, point)
}

func positionDeltaExportCall4FEA70(object *server.Object, point *types.Pointf) int32 {
	return int32(C.sub_4FEA70(
		asObjectC(object),
		(*C.float2)(unsafe.Pointer(point)),
	))
}

//export sub_4FEA70
func sub_4FEA70(object *C.nox_object_t, point *C.float2) C.int32_t {
	return C.int32_t(positionDeltaLegacy4FEA70(
		asObjectS((*nox_object_t)(object)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
