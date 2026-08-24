package ccall

/*
#include <stdint.h>

static uintptr_t go_call_uptr_ptr_int_uptr2_func(
	uintptr_t (*fnc)(void*, int, uintptr_t, uintptr_t),
	void* a1, int a2, uintptr_t a3, uintptr_t a4
) {
	return fnc(a1, a2, a3, a4);
}
*/
import "C"
import "unsafe"

// CallUPtrPtrIntUPtr2 invokes the native-width GUI callback ABI. The event
// number remains a 32-bit C int, while event payloads and responses may carry
// pointers and therefore follow uintptr_t.
func CallUPtrPtrIntUPtr2(fnc unsafe.Pointer, a1 unsafe.Pointer, a2 int, a3, a4 uintptr) uintptr {
	return uintptr(C.go_call_uptr_ptr_int_uptr2_func(
		(*[0]byte)(fnc),
		a1,
		C.int(a2),
		C.uintptr_t(a3),
		C.uintptr_t(a4),
	))
}
