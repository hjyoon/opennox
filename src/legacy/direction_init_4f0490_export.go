package legacy

/*
#include "direction_init_4f0490.h"
*/
import "C"

//export sub_4F0490
func sub_4F0490(unit *C.nox_object_t) C.int32_t {
	return C.int32_t(directionInitCall4F0490(asObjectS((*nox_object_t)(unit))))
}
