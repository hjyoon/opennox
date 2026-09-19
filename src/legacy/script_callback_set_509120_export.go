package legacy

/*
#include "script_callback_set_509120.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var scriptCallbackSetCall509120 = func(object *server.Object, event int, name string) {
	GetServer().S().NoxScriptVM.Nox_script_objCallbackSet_509120(object, event, name)
}

func scriptCallbackSetExportCall509120(object *server.Object, event int32, name string) {
	cname, freeName := alloc.CString(name)
	defer freeName()
	C.sub_509120(
		(*C.nox_object_t)(unsafe.Pointer(object)),
		C.int32_t(event),
		(*C.char)(unsafe.Pointer(cname)),
	)
}

//export sub_509120
func sub_509120(object *C.nox_object_t, event C.int32_t, name *C.char) {
	scriptCallbackSetCall509120(
		asObjectS((*nox_object_t)(object)),
		int(int32(event)),
		GoString(name),
	)
}
