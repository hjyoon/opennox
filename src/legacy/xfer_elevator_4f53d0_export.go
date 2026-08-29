package legacy

/*
#include "xfer_elevator_4f53d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_XFerElevator_4F53D0
func nox_xxx_XFerElevator_4F53D0(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_XFerElevatorNative4F53D0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
