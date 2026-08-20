package legacy

/*
#include "unit_get_old_mana_4eec80.h"
*/
import "C"

//export nox_xxx_unitGetOldMana_4EEC80
func nox_xxx_unitGetOldMana_4EEC80(unit *C.nox_object_t) C.short {
	return C.short(int16(Nox_xxx_unitGetOldMana_4EEC80(asObjectS((*nox_object_t)(unit)))))
}
