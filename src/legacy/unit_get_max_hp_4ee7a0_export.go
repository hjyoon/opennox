package legacy

/*
#include "unit_get_max_hp_4ee7a0.h"
*/
import "C"

//export nox_xxx_unitGetMaxHP_4EE7A0
func nox_xxx_unitGetMaxHP_4EE7A0(unit *C.nox_object_t) C.short {
	return C.short(int16(Nox_xxx_unitGetMaxHP_4EE7A0(asObjectS((*nox_object_t)(unit)))))
}
