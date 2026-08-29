package legacy

/*
#include "xfer_hole_4f51d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_XFerHole_4F51D0
func nox_xxx_XFerHole_4F51D0(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_XFerHoleNative4F51D0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
