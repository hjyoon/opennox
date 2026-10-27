package legacy

/*
#include "GAME5_2.h"
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func pointInBoxCABI57B850(source *types.Pointf, shape *server.Shape, target *types.Pointf) bool {
	return C.nox_xxx_map_57B850((*C.float2)(unsafe.Pointer(source)), (*C.float)(unsafe.Pointer(shape)), (*C.float2)(unsafe.Pointer(target))) != 0
}

func elevatorCollisionCABI551AE0(elevator, candidate *server.Object, move int32) {
	C.sub_551AE0(asObjectC(elevator), asObjectC(candidate), C.int(move))
}
