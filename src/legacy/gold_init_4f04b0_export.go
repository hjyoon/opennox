package legacy

/*
#include "gold_init_4f04b0.h"
*/
import "C"

//export nox_xxx_unitInitGold_4F04B0
func nox_xxx_unitInitGold_4F04B0(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(goldInitCall4F04B0(asObjectS((*nox_object_t)(unit))))
}
