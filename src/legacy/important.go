package legacy

/*
#include "GAME3_3.h"

extern uint32_t dword_5d4594_1565512;
extern uint32_t dword_5d4594_1565516;
*/
import "C"

import "unsafe"

func importantAllocClassC() unsafe.Pointer {
	return unsafe.Pointer(C.nox_server_getImportantAllocClass_4E4DE0())
}

func setImportantAllocClassC(p unsafe.Pointer) {
	C.nox_server_setImportantAllocClass_4E4DE0((*C.nox_alloc_class)(p))
}

func importantListHeadsC() (first, last uint32) {
	return uint32(C.dword_5d4594_1565512), uint32(C.dword_5d4594_1565516)
}

func setImportantListHeadsC(first, last uint32) {
	C.dword_5d4594_1565512 = C.uint32_t(first)
	C.dword_5d4594_1565516 = C.uint32_t(last)
}
