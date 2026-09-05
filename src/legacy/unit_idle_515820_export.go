package legacy

/*
#include "unit_idle_515820.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var unitIdleCall515820 = func(unit *server.Object) {
	GetServer().S().UnitIdle515820(unit)
}

func unitIdleExportCall515820(unit *server.Object) {
	C.nox_xxx_unitIdle_515820(asObjectC(unit))
}

//export nox_xxx_unitIdle_515820
func nox_xxx_unitIdle_515820(unit *C.nox_object_t) {
	unitIdleCall515820(asObjectS((*nox_object_t)(unit)))
}
