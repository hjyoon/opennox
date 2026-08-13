package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export nox_xxx_doorGetSomeKey_4E8910
func nox_xxx_doorGetSomeKey_4E8910(unit, door *C.nox_object_t) *C.nox_object_t {
	return (*C.nox_object_t)(asObjectC(GetServer().S().DoorCheckKey(
		asObjectS((*nox_object_t)(unit)),
		asObjectS((*nox_object_t)(door)),
	)))
}
