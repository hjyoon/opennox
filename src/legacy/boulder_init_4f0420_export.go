package legacy

/*
#include "boulder_init_4f0420.h"
*/
import "C"

//export nox_xxx_unitBoulderInit_4F0420
func nox_xxx_unitBoulderInit_4F0420(unit *C.nox_object_t) *C.nox_object_t {
	result := boulderInitCall4F0420(asObjectS((*nox_object_t)(unit)))
	return (*C.nox_object_t)(asObjectC(result))
}
