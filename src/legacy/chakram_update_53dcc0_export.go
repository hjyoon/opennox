package legacy

/*
#include "chakram_update_53dcc0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_updateChakramInMotion_53DCC0
func nox_xxx_updateChakramInMotion_53DCC0(source *C.nox_object_t) {
	srv := GetServer()
	srv.S().ChakramInMotionUpdate53DCC0(
		asObjectS((*nox_object_t)(source)),
		server.ChakramInMotionUpdateRuntime53DCC0{
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
