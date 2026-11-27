package ccall

/*
#include "ccall_abi.h"
*/
import "C"
import "unsafe"

// CallWindowProc invokes the signed-int GUI event ABI.
func CallWindowProc(fnc, win unsafe.Pointer, event int, a, b uintptr) int {
	return int(C.nox_ccall_window_proc((*[0]byte)(fnc), (*C.nox_window)(win),
		C.int(event), C.uintptr_t(a), C.uintptr_t(b)))
}

// CallUPtrPtrIntUPtr2 invokes a window initializer with a native-width result.
func CallUPtrPtrIntUPtr2(fnc unsafe.Pointer, a1 unsafe.Pointer, a2 int, a3, a4 uintptr) uintptr {
	return uintptr(C.nox_ccall_uptr_ptr_int_uptr2(
		(*[0]byte)(fnc),
		(*C.nox_window)(a1),
		C.int(a2),
		C.uintptr_t(a3),
		C.uintptr_t(a4),
	))
}

func CallWindowDraw(fnc, win, draw unsafe.Pointer) int {
	return int(C.nox_ccall_window_draw((*[0]byte)(fnc), (*C.nox_window)(win), (*C.nox_window_data)(draw)))
}

func CallWindowTooltip(fnc, win, draw unsafe.Pointer, pos uintptr) {
	C.nox_ccall_window_tooltip((*[0]byte)(fnc), (*C.nox_window)(win), (*C.nox_window_data)(draw), C.int(pos))
}
