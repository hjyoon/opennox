package legacy

/*
#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerActionStateExportCall4FA2B0(unit *server.Object) int32 {
	return int32(C.nox_common_mapPlrActionToStateId_4FA2B0(asObjectC(unit)))
}

//export nox_common_mapPlrActionToStateId_4FA2B0
func nox_common_mapPlrActionToStateId_4FA2B0(unit *C.nox_object_t) C.int {
	return C.int(GetServer().S().PlayerActionState4FA2B0(
		asObjectS((*nox_object_t)(unit)),
	))
}
