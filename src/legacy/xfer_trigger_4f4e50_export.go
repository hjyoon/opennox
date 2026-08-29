package legacy

/*
#include "xfer_trigger_4f4e50.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_unitTriggerXfer_4F4E50
func nox_xxx_unitTriggerXfer_4F4E50(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_UnitTriggerXferNative4F4E50(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
