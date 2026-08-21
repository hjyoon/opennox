package legacy

/*
#include "chest_init_4f0400.h"
*/
import "C"

//export nox_xxx_initChest_4F0400
func nox_xxx_initChest_4F0400(unit *C.nox_object_t) {
	chestInitCall4F0400(asObjectS((*nox_object_t)(unit)))
}
