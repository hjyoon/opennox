package legacy

/*
#include "unit_set_max_hp_4ee7c0.h"
*/
import "C"

import "unsafe"

//export nox_xxx_unitSetMaxHP_4EE7C0
func nox_xxx_unitSetMaxHP_4EE7C0(unit *C.nox_object_t, maximum C.short) unsafe.Pointer {
	return unsafe.Pointer(Nox_xxx_unitSetMaxHP_4EE7C0(
		asObjectS((*nox_object_t)(unit)),
		uint16(maximum),
	))
}
