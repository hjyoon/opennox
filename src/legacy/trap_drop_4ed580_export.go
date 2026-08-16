package legacy

/*
#include "trap_drop_4ed580.h"

int nox_xxx_mapTileAllowTeleport_411A90(float2* point);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

func mapTileAllowTeleport411A90(point *types.Pointf) int32 {
	return int32(C.nox_xxx_mapTileAllowTeleport_411A90((*C.float2)(unsafe.Pointer(point))))
}

//export nox_xxx_dropTrap_4ED580
func nox_xxx_dropTrap_4ED580(
	owner, trap *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(trapDropCall4ED580(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(trap)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
