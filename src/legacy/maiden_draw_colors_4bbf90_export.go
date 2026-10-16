package legacy

/*
#include "client__draw__maidendraw.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_client_maidenColors_native_4BBF90
func nox_client_maidenColors_native_4BBF90(obj *C.nox_object_t) *C.nox_color3_t {
	// The original selects by extent and reads data_update directly. Do not
	// introduce a class/nil guard or copy the array: material callbacks may
	// replace data_update and change later colors in the cached record.
	ud := (*server.MonsterUpdateData)(asObjectS(obj).UpdateData)
	return (*C.nox_color3_t)(unsafe.Pointer(&ud.Color[0]))
}
