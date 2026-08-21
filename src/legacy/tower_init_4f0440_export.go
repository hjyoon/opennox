package legacy

/*
#include "tower_init_4f0440.h"
*/
import "C"

//export nox_xxx_unitTowerInit_4F0440
func nox_xxx_unitTowerInit_4F0440(unit *C.nox_object_t) {
	towerInitCall4F0440(asObjectS((*nox_object_t)(unit)))
}
