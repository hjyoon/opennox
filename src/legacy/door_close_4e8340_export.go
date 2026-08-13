package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_fnFindCloseDoors_4E8340
func nox_xxx_fnFindCloseDoors_4E8340(door *nox_object_t, target *C.nox_point) {
	GetServer().S().DoorCloseAtTile4E8340(
		asObjectS(door),
		(*server.DoorTilePoint)(unsafe.Pointer(target)),
	)
}

//export sub_4E8390
func sub_4E8390(door *nox_object_t) int32 {
	return GetServer().S().DoorQuestSync4E8390(asObjectS(door))
}

//export sub_4D6A20
func sub_4D6A20(recipient int32, object *nox_object_t) int32 {
	return GetServer().S().DoorExtentPacket4D6A20(recipient, asObjectS(object))
}
