package legacy

/*
#include "GAME4_2.h"
*/
import "C"
import "unsafe"

//export nox_xxx_castFireball_52C790
func nox_xxx_castFireball_52C790(id C.int, second unsafe.Pointer, caster *C.nox_object_t, firing *C.nox_object_t, arg unsafe.Pointer, level C.int) C.int {
	return C.int(fireballCall52C790(int32(id), asObjectS((*nox_object_t)(firing)), int32(level)))
}
