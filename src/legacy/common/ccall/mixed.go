package ccall

/*
#include <stdint.h>

typedef struct nox_window nox_window;

static int go_call_window_proc(
	int (*fnc)(nox_window*, int, uintptr_t, uintptr_t),
	nox_window* win, int event, uintptr_t a, uintptr_t b
) {
	return fnc(win, event, a, b);
}

static uintptr_t go_call_uptr_ptr_int_uptr2_func(
	uintptr_t (*fnc)(nox_window*, int, uintptr_t, uintptr_t),
	nox_window* a1, int a2, uintptr_t a3, uintptr_t a4
) {
	return fnc(a1, a2, a3, a4);
}
*/
import "C"
import "unsafe"

// CallWindowProc invokes the signed-int GUI event ABI.
func CallWindowProc(fnc, win unsafe.Pointer, event int, a, b uintptr) int {
	return int(C.go_call_window_proc((*[0]byte)(fnc), (*C.nox_window)(win),
		C.int(event), C.uintptr_t(a), C.uintptr_t(b)))
}

// CallUPtrPtrIntUPtr2 invokes a window initializer with a native-width result.
func CallUPtrPtrIntUPtr2(fnc unsafe.Pointer, a1 unsafe.Pointer, a2 int, a3, a4 uintptr) uintptr {
	return uintptr(C.go_call_uptr_ptr_int_uptr2_func(
		(*[0]byte)(fnc),
		(*C.nox_window)(a1),
		C.int(a2),
		C.uintptr_t(a3),
		C.uintptr_t(a4),
	))
}
