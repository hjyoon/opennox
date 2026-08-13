package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func objectSetOnRuntime4E75B0(obj *server.Object) byte {
	return objectSetOnNative4E75B0(
		obj,
		func(obj *server.Object) {
			nox_xxx_aud_501960(235, asObjectC(obj), 0, 0)
		},
		func(obj *server.Object) byte {
			return byte(C.nox_xxx_unitHasCollideOrUpdateFn_537610(asObjectC(obj)))
		},
	)
}
