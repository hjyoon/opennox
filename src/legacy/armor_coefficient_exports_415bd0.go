package legacy

/*
#include "GAME1.h"
*/
import "C"

import "unsafe"

//export sub_415BD0
func sub_415BD0(item *C.nox_object_t) C.double {
	return C.double(armorCoefficientNative415BD0(unsafe.Pointer(item)))
}
