package legacy

/*
#include "unit_hp_set_max_4ee6f0.h"
*/
import "C"

//export nox_xxx_unitHPsetOnMax_4EE6F0
func nox_xxx_unitHPsetOnMax_4EE6F0(unit *C.nox_object_t) {
	Nox_xxx_unitHPsetOnMax_4EE6F0(asObjectS((*nox_object_t)(unit)))
}
