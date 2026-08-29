package legacy

/*
#include "xfer_door_4f4cb0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_XFerDoor_4F4CB0
func nox_xxx_XFerDoor_4F4CB0(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_XFerDoorNative4F4CB0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
