package legacy

/*
#include "skull_init_4f0450.h"
*/
import "C"

//export nox_xxx_unitSkullInit_4F0450
func nox_xxx_unitSkullInit_4F0450(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(skullInitCall4F0450(asObjectS((*nox_object_t)(unit))))
}
