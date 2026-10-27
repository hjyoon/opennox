package legacy

/*
#include "GAME5_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_map_57B850_go
func nox_xxx_map_57B850_go(source *C.float2, shape *C.float, target *C.float2) C.int {
	return C.int(bool2int(server.MapPointInBox57B850((*types.Pointf)(unsafe.Pointer(source)), (*server.Shape)(unsafe.Pointer(shape)), (*types.Pointf)(unsafe.Pointer(target)))))
}
