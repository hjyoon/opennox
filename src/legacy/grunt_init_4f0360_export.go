package legacy

/*
#include "grunt_init_4f0360.h"
*/
import "C"

//export nox_xxx_unitGruntInit_4F0360
func nox_xxx_unitGruntInit_4F0360(unit *C.nox_object_t) {
	gruntInitCall4F0360(asObjectS((*nox_object_t)(unit)))
}
