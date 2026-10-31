package legacy

/*
#include "defs.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var objectDispatchRuntimeNative519410 = func() server.ObjectDispatchNativeRuntime519410 {
	return GetServer().ObjectDispatchRuntimeNative519410()
}

func NetObjectDispatch519410(recipient, unit *server.Object) byte {
	result, _ := server.ObjectDispatchNative519410(recipient, unit, objectDispatchRuntimeNative519410())
	return result
}

//export nox_xxx_netSendObjects2Plr_519410
func nox_xxx_netSendObjects2Plr_519410(recipient, unit *C.nox_object_t) C.char {
	return C.char(NetObjectDispatch519410(asObjectS(recipient), asObjectS(unit)))
}
