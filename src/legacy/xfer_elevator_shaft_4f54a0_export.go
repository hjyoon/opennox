package legacy

/*
#include "xfer_elevator_shaft_4f54a0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var elevatorShaftXferCall4F54A0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
	_ unsafe.Pointer,
) int32 {
	return Nox_xxx_XFerElevatorShaftNative4F54A0(cf, object)
}

func elevatorShaftXferExportCall4F54A0(
	object *server.Object,
	context unsafe.Pointer,
) int32 {
	return int32(C.nox_xxx_XFerElevatorShaft_4F54A0(asObjectC(object), context))
}

//export nox_xxx_XFerElevatorShaft_4F54A0
func nox_xxx_XFerElevatorShaft_4F54A0(
	object *C.nox_object_t,
	context unsafe.Pointer,
) C.int32_t {
	return C.int32_t(elevatorShaftXferCall4F54A0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
		context,
	))
}
