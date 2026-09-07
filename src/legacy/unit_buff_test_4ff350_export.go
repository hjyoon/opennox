package legacy

/*
#include "unit_buff_test_4ff350.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func unitBuffTestLegacy4FF350(unit *server.Object, buff int32) int32 {
	return unit.UnitBuffTest4FF350(buff)
}

func unitBuffTestExportCall4FF350(unit *server.Object, buff int32) int32 {
	return int32(C.nox_xxx_testUnitBuffs_4FF350(asObjectC(unit), C.int32_t(buff)))
}

//export nox_xxx_testUnitBuffs_4FF350
func nox_xxx_testUnitBuffs_4FF350(unit *C.nox_object_t, buff C.int32_t) C.int32_t {
	return C.int32_t(unitBuffTestLegacy4FF350(
		asObjectS((*nox_object_t)(unit)),
		int32(buff),
	))
}
