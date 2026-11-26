//go:build !386 && !amd64 && !arm && !arm64

package legacy

// #include "noxstring_format_abi.h"
import "C"

import "unsafe"

func nativeFormatInteger(args *nativeFormatArgs, unsigned bool) int32 {
	if unsigned {
		return int32(C.nox_format_arg_uint(args))
	}
	return int32(C.nox_format_arg_int(args))
}

func nativeFormatStringPointer(args *nativeFormatArgs, wide bool) unsafe.Pointer {
	if wide {
		return unsafe.Pointer(C.nox_format_arg_wide(args))
	}
	return unsafe.Pointer(C.nox_format_arg_narrow(args))
}

func nativeFormatDoubleBits(args *nativeFormatArgs) uint64 {
	return uint64(C.nox_format_arg_double_bits(args))
}
