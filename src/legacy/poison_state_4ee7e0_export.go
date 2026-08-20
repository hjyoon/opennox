package legacy

/*
#include "poison_state_4ee7e0.h"
*/
import "C"

//export nox_xxx_activatePoison_4EE7E0
func nox_xxx_activatePoison_4EE7E0(unit *C.nox_object_t, increment, maximum C.int32_t) C.int32_t {
	return C.int32_t(Nox_xxx_activatePoison_4EE7E0(
		asObjectS((*nox_object_t)(unit)),
		int32(increment),
		int32(maximum),
	))
}

//export nox_xxx_updatePoison_4EE8F0
func nox_xxx_updatePoison_4EE8F0(unit *C.nox_object_t, amount C.int32_t) {
	GetServer().S().UpdatePoison4EE8F0(asObjectS((*nox_object_t)(unit)), int32(amount))
}

//export nox_xxx_removePoison_4EE9D0
func nox_xxx_removePoison_4EE9D0(unit *C.nox_object_t) {
	GetServer().S().RemovePoison4EE9D0(asObjectS((*nox_object_t)(unit)))
}

//export nox_xxx_setSomePoisonData_4EEA90
func nox_xxx_setSomePoisonData_4EEA90(unit *C.nox_object_t, value C.int32_t) {
	setSomePoisonDataCall4EEA90(asObjectS((*nox_object_t)(unit)), int32(value))
}
