//go:build !server

package legacy

/*
#include "defs.h"
#include "GAME2_2.h"
extern uint32_t dword_5d4594_3804684;

// Follow the entry's actual declaration on either side of the port. Its
// parameter and result have the same type, so this probe compiles against the
// original int(int) and the corrected void*(void*) without an incompatible
// function-pointer call. No callback or blob accessor is substituted.
static uintptr_t nox_test_video_setter_return_48A210(void* callback) {
	return (uintptr_t)nox_xxx_setSomeFunc_48A210(
		(__typeof__(nox_xxx_setSomeFunc_48A210(0)))(uintptr_t)callback);
}
static int nox_test_video_setter_native_declaration_48A210(void) {
	return _Generic(&nox_xxx_setSomeFunc_48A210,
		void* (*)(void*): 1, default: 0);
}
static int nox_test_video_callback_call_48A210(uint32_t* descriptor, int ignored) {
	typedef int (*callback_t)(uint32_t*, int);
	void* callback = getMemPtr(0x5D4594, 1193504);
	if (!callback) {
		return INT32_MIN;
	}
	return ((callback_t)callback)(descriptor, ignored);
}
static int nox_test_video_callback_direct_47FCE0(uint32_t* descriptor, int ignored) {
	return sub_47FCE0(descriptor, ignored);
}
*/
import "C"
import "unsafe"

// Declaration-only alias to the real descriptor counter.
var videoCallbackDescriptorCountC48A210 = (*uint32)(unsafe.Pointer(&C.dword_5d4594_3804684))

func videoCallbackSetterReturnC48A210(callback unsafe.Pointer) uintptr {
	return uintptr(C.nox_test_video_setter_return_48A210(callback))
}
func videoCallbackSetterNativeDeclaration48A210() bool {
	return C.nox_test_video_setter_native_declaration_48A210() != 0
}
func videoCallbackStoredCall48A210(descriptor *uint32, ignored uint32) int32 {
	return int32(C.nox_test_video_callback_call_48A210((*C.uint32_t)(descriptor), C.int(ignored)))
}
func videoCallbackDirectCall47FCE0(descriptor *uint32, ignored uint32) int32 {
	return int32(C.nox_test_video_callback_direct_47FCE0((*C.uint32_t)(descriptor), C.int(ignored)))
}
