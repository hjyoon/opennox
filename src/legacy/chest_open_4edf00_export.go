package legacy

/*
#include "chest_open_4edf00.h"
*/
import "C"

//export nox_xxx_chest_4EDF00
func nox_xxx_chest_4EDF00(chest, unit *C.nox_object_t) {
	chestOpenCall4EDF00(
		asObjectS((*nox_object_t)(chest)),
		asObjectS((*nox_object_t)(unit)),
	)
}
