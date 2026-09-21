package legacy

/*
#include "GAME1_3.h"
*/
import "C"

import "unsafe"

func SetXferBuffer446520(ind int, ptr unsafe.Pointer) {
	C.nox_xxx_xferSetBuffer_446520(C.int(ind), ptr)
}

func GetXferBuffer446520(ind int) unsafe.Pointer {
	return C.nox_xxx_xferGetBuffer_446520(C.int(ind))
}
