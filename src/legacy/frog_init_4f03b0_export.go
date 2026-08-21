package legacy

/*
#include "frog_init_4f03b0.h"
*/
import "C"

//export nox_xxx_initFrog_4F03B0
func nox_xxx_initFrog_4F03B0(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(frogInitCall4F03B0(asObjectS((*nox_object_t)(unit))))
}
