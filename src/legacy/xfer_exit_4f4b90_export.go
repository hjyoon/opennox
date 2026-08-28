package legacy

/*
#include "xfer_exit_4f4b90.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_XFerExit_4F4B90
func nox_xxx_XFerExit_4F4B90(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_XFerExitNative4F4B90(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
