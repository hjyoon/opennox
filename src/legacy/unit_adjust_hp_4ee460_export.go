package legacy

/*
#include "unit_adjust_hp_4ee460.h"
*/
import "C"

//export nox_xxx_unitAdjustHP_4EE460
func nox_xxx_unitAdjustHP_4EE460(unit *C.nox_object_t, delta C.int) {
	unitAdjustHPCall4EE460(
		asObjectS((*nox_object_t)(unit)),
		int32(delta),
	)
}

//export nox_xxx_mobInformOwnerHP_4EE4C0
func nox_xxx_mobInformOwnerHP_4EE4C0(obj *C.nox_object_t) {
	mobInformOwnerHPCall4EE4C0(asObjectS((*nox_object_t)(obj)))
}

//export nox_xxx_netReportUnitCurrentHP_4D8620
func nox_xxx_netReportUnitCurrentHP_4D8620(recipient C.int, obj *C.nox_object_t) C.int {
	return C.int(currentHPReportCall4D8620(
		int32(recipient),
		asObjectS((*nox_object_t)(obj)),
	))
}
