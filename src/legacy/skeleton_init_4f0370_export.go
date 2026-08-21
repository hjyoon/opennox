package legacy

/*
#include "skeleton_init_4f0370.h"
*/
import "C"

//export nox_xxx_unitSkeletonInit_4F0370
func nox_xxx_unitSkeletonInit_4F0370(unit *C.nox_object_t) {
	skeletonInitCall4F0370(asObjectS((*nox_object_t)(unit)))
}
