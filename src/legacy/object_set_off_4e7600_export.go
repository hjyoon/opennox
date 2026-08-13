package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func objectSetOffRuntime4E7600(obj *server.Object) uint32 {
	return objectSetOffNative4E7600(obj, func(obj *server.Object) {
		nox_xxx_aud_501960(236, asObjectC(obj), 0, 0)
	})
}

//export nox_xxx_objectSetOff_4E7600
func nox_xxx_objectSetOff_4E7600(obj *nox_object_t) C.int {
	return C.int(int32(objectSetOffRuntime4E7600(asObjectS(obj))))
}
