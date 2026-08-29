package legacy

/*
#include "defs.h"
#include "script_callback_init_4f5540.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_4F5540
func sub_4F5540(handler *C.nox_script_callback_t) C.int32_t {
	return C.int32_t(scriptCallbackInitRuntime4F5540(
		(*server.ScriptCallback)(unsafe.Pointer(handler)),
	))
}
