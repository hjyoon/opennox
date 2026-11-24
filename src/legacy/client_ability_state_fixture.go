package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import "unsafe"

func clientAbilityActive461160(id int) bool {
	return C.sub_461160(C.int(id)) != 0
}

func clientAbilityState461090(id int, state uint32) uintptr {
	return uintptr(unsafe.Pointer(C.sub_461090(C.int(id), C.int(state))))
}

func clientAbilityReset4610D0(index int) uintptr {
	return uintptr(unsafe.Pointer(C.sub_4610D0(C.uchar(index))))
}

func clientAbilitySetActive461120(id, state int) uintptr {
	return uintptr(unsafe.Pointer(C.sub_461120(C.int(id), C.int(state))))
}
