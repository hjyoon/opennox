package legacy

/*
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export sub_410390
func sub_410390(drawable *C.nox_drawable, x, y C.int) unsafe.Pointer {
	return sub410390Native(unsafe.Pointer(drawable), int32(x), int32(y))
}
