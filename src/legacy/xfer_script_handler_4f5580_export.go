package legacy

/*
#include "defs.h"
#include "server__script__script.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var scriptHandlerXferCall4F5580 = scriptHandlerXferRuntime4F5580

func scriptHandlerXferExportCall4F5580(
	handler *server.ScriptCallback,
	context unsafe.Pointer,
) int32 {
	return int32(C.nox_xxx_xferReadScriptHandler_4F5580(
		(*C.nox_script_callback_t)(unsafe.Pointer(handler)),
		(*C.char)(context),
	))
}

//export nox_xxx_xferReadScriptHandler_4F5580
func nox_xxx_xferReadScriptHandler_4F5580(
	handler *C.nox_script_callback_t,
	context *C.char,
) C.int32_t {
	return C.int32_t(scriptHandlerXferCall4F5580(
		(*server.ScriptCallback)(unsafe.Pointer(handler)),
		unsafe.Pointer(context),
	))
}
