package legacy

/*
#include "creature_monitored_500cc0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var creatureMonitoredCall500CC0 = server.Nox_xxx_creatureIsMonitored_500CC0

func creatureMonitoredExportCall500CC0(owner, unit *server.Object) int32 {
	return int32(C.nox_xxx_creatureIsMonitored_500CC0(
		asObjectC(owner),
		asObjectC(unit),
	))
}

//export nox_xxx_creatureIsMonitored_500CC0
func nox_xxx_creatureIsMonitored_500CC0(owner, unit *C.nox_object_t) C.int32_t {
	return C.int32_t(bool2int(creatureMonitoredCall500CC0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(unit)),
	)))
}
