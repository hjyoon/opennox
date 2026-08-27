package legacy

/*
#include "GAME2_3.h"

extern void* dword_5d4594_1203864;
*/
import "C"

import "unsafe"

func Nox_xxx_cliAddObjFriend_4959F0(netCode uint32) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_cliAddObjFriend_4959F0(C.int(netCode)))
}

func Sub_4959B0() {
	C.sub_4959B0()
}

func Sub_495A20(netCode uint32) {
	C.sub_495A20(C.int(netCode))
}

func friendListHead495980() unsafe.Pointer {
	return C.dword_5d4594_1203864
}
