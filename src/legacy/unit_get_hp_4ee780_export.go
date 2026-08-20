package legacy

/*
#include "unit_get_hp_4ee780.h"
*/
import "C"

//export nox_xxx_unitGetHP_4EE780
func nox_xxx_unitGetHP_4EE780(unit *C.nox_object_t) C.short {
	return C.short(int16(Nox_xxx_unitGetHP_4EE780(asObjectS((*nox_object_t)(unit)))))
}
