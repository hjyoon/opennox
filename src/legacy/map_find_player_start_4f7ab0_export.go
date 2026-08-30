package legacy

/*
#include "map_find_player_start_4f7ab0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func mapFindPlayerStartExportCall4F7AB0(output *types.Pointf, player *server.Object) {
	C.nox_xxx_mapFindPlayerStart_4F7AB0(
		(*C.float2)(unsafe.Pointer(output)),
		asObjectC(player),
	)
}

//export nox_xxx_mapFindPlayerStart_4F7AB0
func nox_xxx_mapFindPlayerStart_4F7AB0(output *C.float2, player *C.nox_object_t) {
	GetServer().S().MapFindPlayerStartInto4F7AB0(
		(*types.Pointf)(unsafe.Pointer(output)),
		asObjectS((*nox_object_t)(player)),
	)
}
