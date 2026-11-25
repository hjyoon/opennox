package legacy

/*
#include "noxstring_format_abi.h"

void nox_format_wide_strings_abi(wchar2_t* out, size_t capacity, const wchar2_t* format,
                               const wchar2_t* first, const wchar2_t* second);
*/
import "C"
import "unsafe"

func formatWideStrings(out []wchar2_t, format, first, second *wchar2_t) {
	C.nox_format_wide_strings_abi(&out[0], C.size_t(len(out)), format, first, second)
}

//export nox_format_bounded_wide_go
func nox_format_bounded_wide_go(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, args *C.nox_format_args) {
	out := unsafe.Slice(dst, uintptr(capacity))
	nox_format_wide_go(dst, capacity-1, format, args)
	out[len(out)-1] = 0
}
