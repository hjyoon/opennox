package legacy

/*
#include "decay_511660.h"
*/
import "C"

//export nox_xxx_unitSetDecayTime_511660
func nox_xxx_unitSetDecayTime_511660(obj *C.nox_object_t, delay C.int) C.int {
	result := GetServer().S().DecaySetTime511660(
		asObjectS((*nox_object_t)(obj)),
		uint32(delay),
	)
	return C.int(result)
}

//export nox_xxx_decay_5116F0
func nox_xxx_decay_5116F0(obj *C.nox_object_t) {
	GetServer().S().DecayRemove5116F0(asObjectS((*nox_object_t)(obj)))
}
