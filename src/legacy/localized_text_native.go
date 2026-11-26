package legacy

/*
#include "noxstring_format_abi.h"
*/
import "C"
import "unsafe"

func formatWideNumber(out []wchar2_t, format *wchar2_t, value C.int) {
	values := [...]formatValue{{number: int32(value)}}
	formatBoundedWide(out, format, formatArguments{values: values[:]})
}

func formatWideStrings(out []wchar2_t, format, first, second *wchar2_t) {
	values := [...]formatValue{{text: unsafe.Pointer(first)}, {text: unsafe.Pointer(second)}}
	formatBoundedWide(out, format, formatArguments{values: values[:]})
}

func formatFixedWide(dst, format *wchar2_t, values ...formatValue) C.int {
	// Retain the nox_vswprintf capacity and its result in code units.
	return C.int(formatWithArgs(
		(*uint16)(unsafe.Pointer(dst)), 0x3fffffff,
		(*uint16)(unsafe.Pointer(format)), formatArguments{values: values},
	))
}

func formatBoundedWide(out []wchar2_t, format *wchar2_t, args formatArguments) {
	if len(out) == 0 {
		return
	}
	formatWithArgs((*uint16)(unsafe.Pointer(&out[0])), uintptr(len(out)-1), (*uint16)(unsafe.Pointer(format)), args)
	out[len(out)-1] = 0
}

//export nox_format_wide_number_abi
func nox_format_wide_number_abi(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, value C.int) {
	formatWideNumber(unsafe.Slice(dst, capacity), format, value)
}

//export nox_format_wide_strings_abi
func nox_format_wide_strings_abi(dst *C.wchar2_t, capacity C.size_t, format, first, second *C.nox_format_word) {
	formatWideStrings(unsafe.Slice(dst, capacity), format, first, second)
}

//export nox_format_bounded_wide_go
func nox_format_bounded_wide_go(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, args *C.nox_format_args) {
	formatBoundedWide(unsafe.Slice(dst, capacity), format, formatArguments{native: args})
}
