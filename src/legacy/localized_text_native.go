package legacy

/*
#include "noxstring_format_abi.h"
*/
import "C"
import "unsafe"

func formatWideNumber(out []wchar2_t, format *wchar2_t, value C.int) {
	C.nox_format_wide_number_abi(&out[0], C.size_t(len(out)), format, value)
}

func formatWideStrings(out []wchar2_t, format, first, second *wchar2_t) {
	C.nox_format_wide_strings_abi(&out[0], C.size_t(len(out)), format, first, second)
}

//export nox_format_bounded_wide_go
func nox_format_bounded_wide_go(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, args *C.nox_format_args) {
	out := unsafe.Slice(dst, uintptr(capacity))
	nox_format_wide_go(dst, capacity-1, format, args)
	out[len(out)-1] = 0
}
