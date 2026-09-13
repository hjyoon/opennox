package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var waterBarrelUpdateCall53CB90 = func(source *server.Object) {
	outer := GetServer()
	outer.S().WaterBarrelUpdate53CB90(source, server.WaterBarrelUpdateRuntime53CB90{
		DelayedDelete: outer.DelayedDelete,
	})
}

func waterBarrelUpdateExportCall53CB90(source *server.Object) {
	C.nox_xxx_updateWaterBarrel_53CB90((*C.nox_object_t)(source.CObj()))
}

//export nox_xxx_updateWaterBarrel_53CB90
func nox_xxx_updateWaterBarrel_53CB90(source *C.nox_object_t) {
	waterBarrelUpdateCall53CB90(asObjectS((*nox_object_t)(source)))
}
