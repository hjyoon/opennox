package legacy

/*
#include "object_read_old_4f4170.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_readObjectOldVer_4F4170
func nox_xxx_readObjectOldVer_4F4170(
	object *C.nox_object_t,
	objectVersion, mapVersion C.int32_t,
) C.int32_t {
	return C.int32_t(objectReadOldNative4F4170(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
		int32(objectVersion),
		int32(mapVersion),
	))
}
