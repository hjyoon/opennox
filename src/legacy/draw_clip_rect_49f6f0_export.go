package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

// Set during video initialization. The video package owns both render-data buffers.
var currentDrawData *C.nox_render_data_t

func currentDrawClipRect(x, y, w, h C.int) C.int {
	if drawClipRectNative49F6F0((*noxrender.RenderData)(unsafe.Pointer(currentDrawData)), int32(x), int32(y), int32(w), int32(h)) {
		return 1
	}
	return 0
}

//export nox_client_copyRect_current_49F6F0
func nox_client_copyRect_current_49F6F0(x, y, w, h C.int) C.int {
	return currentDrawClipRect(x, y, w, h)
}
