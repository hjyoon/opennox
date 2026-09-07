package legacy

/*
#include "unit_buff_power_4ff570.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBuffPowerLegacy4FF570(unit *server.Object, buff int32) uint8 {
	return unit.UnitBuffPower4FF570(buff)
}

var unitBuffPowerExportImpl4FF570 = unitBuffPowerLegacy4FF570

func unitBuffPowerExportCall4FF570(unit *server.Object, buff int32) uint8 {
	return uint8(C.nox_xxx_buffGetPower_4FF570(
		asObjectC(unit),
		C.int32_t(buff),
	))
}

//export nox_xxx_buffGetPower_4FF570
func nox_xxx_buffGetPower_4FF570(unit *C.nox_object_t, buff C.int32_t) C.uint8_t {
	return C.uint8_t(unitBuffPowerExportImpl4FF570(
		asObjectS((*nox_object_t)(unit)),
		int32(buff),
	))
}
