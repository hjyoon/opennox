package legacy

/*
#include "chakram_update_53dcc0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func chakramInMotionUpdateNative53DCC0(source *server.Object) {
	srv := GetServer()
	srv.S().ChakramInMotionUpdate53DCC0(
		source,
		server.ChakramInMotionUpdateRuntime53DCC0{
			DelayedDelete: srv.DelayedDelete,
		},
	)
}

var chakramInMotionUpdateCall53DCC0 = chakramInMotionUpdateNative53DCC0

//export nox_xxx_updateChakramInMotion_53DCC0
func nox_xxx_updateChakramInMotion_53DCC0(source *C.nox_object_t) {
	chakramInMotionUpdateNative53DCC0(asObjectS((*nox_object_t)(source)))
}
