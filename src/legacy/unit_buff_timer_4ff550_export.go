package legacy

/*
#include "unit_buff_timer_4ff550.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBuffTimerLegacy4FF550(unit *server.Object, buff int32) uint32 {
	return unit.UnitBuffTimer4FF550(buff)
}

var unitBuffTimerExportImpl4FF550 = unitBuffTimerLegacy4FF550

func unitBuffTimerExportCall4FF550(unit *server.Object, buff int32) uint32 {
	return uint32(C.nox_xxx_unitGetBuffTimer_4FF550(
		asObjectC(unit),
		C.int32_t(buff),
	))
}

//export nox_xxx_unitGetBuffTimer_4FF550
func nox_xxx_unitGetBuffTimer_4FF550(unit *C.nox_object_t, buff C.int32_t) C.uint32_t {
	return C.uint32_t(unitBuffTimerExportImpl4FF550(
		asObjectS((*nox_object_t)(unit)),
		int32(buff),
	))
}
