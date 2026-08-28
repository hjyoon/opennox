package legacy

/*
#include "object_map_read_write_4f4530.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_mapReadWriteObjData_4F4530
func nox_xxx_mapReadWriteObjData_4F4530(
	object *C.nox_object_t,
	mapVersion C.int32_t,
) C.int32_t {
	return C.int32_t(objectMapReadWriteNative4F4530(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
		int32(mapVersion),
	))
}
