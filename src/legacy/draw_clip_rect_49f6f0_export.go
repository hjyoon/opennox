package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

//export nox_client_copyRect_native_49F6F0
func nox_client_copyRect_native_49F6F0(data *C.nox_render_data_t, x, y, w, h C.int) C.int {
	if drawClipRectNative49F6F0((*noxrender.RenderData)(unsafe.Pointer(data)), int32(x), int32(y), int32(w), int32(h)) {
		return 1
	}
	return 0
}
