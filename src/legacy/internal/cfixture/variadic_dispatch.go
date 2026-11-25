package cfixture

/*
#include "GAME3_2.h"
#include "GAME4_3.h"

static intptr_t line_message_fixture(nox_object_t* unit, wchar2_t* format, int points) {
	const wchar2_t suffix[] = {'!', 0};
	return nox_xxx_netSendLineMessage_4D9EB0(unit, format, "xp", points, suffix);
}

static void ai_debug_printf_fixture(char* format, int frame) {
	nox_ai_debug_printf_5341A0(format, frame, "monster");
}
*/
import "C"

import "unsafe"

func LineMessage4D9EB0(unit unsafe.Pointer, format *uint16, points int) uintptr {
	return uintptr(C.line_message_fixture((*C.nox_object_t)(unit), (*C.wchar2_t)(unsafe.Pointer(format)), C.int(points)))
}

func AIDebugPrintf5341A0(format *byte, frame int) {
	C.ai_debug_printf_fixture((*C.char)(unsafe.Pointer(format)), C.int(frame))
}
