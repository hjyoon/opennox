//go:build 386 || amd64 || arm || arm64

package legacy

/*
#include <stdio.h>
#include "noxstring_format_abi.h"

#if defined(__x86_64__) && !defined(_WIN32)
// GCC can give cgo a different struct tag for its built-in va_list.
// The adjusted CRT parameter is one cursor pointer on SysV AMD64.
typedef void* nox_crt_format_list_arg;
#define NOX_VPRINTF_STRINGIFY_(x) #x
#define NOX_VPRINTF_STRINGIFY(x) NOX_VPRINTF_STRINGIFY_(x)
int nox_crt_vsnprintf(char*, size_t, const char*, nox_crt_format_list_arg)
	__asm__(NOX_VPRINTF_STRINGIFY(__USER_LABEL_PREFIX__) "vsnprintf");
#undef NOX_VPRINTF_STRINGIFY
#undef NOX_VPRINTF_STRINGIFY_
#else
typedef nox_format_list_arg nox_crt_format_list_arg;
#define nox_crt_vsnprintf vsnprintf
#endif
*/
import "C"

import (
	"runtime"
	"unsafe"
)

const sysVFormatListByPointer = runtime.GOARCH == "amd64" && runtime.GOOS != "windows"

// These ABIs copy the cursor by value and have no va_end work after the
// copy. On SysV AMD64 the incoming parameter points to the cursor array.
// Its register-save and stack storage stays borrowed through the call.
func copyNativeFormatArgs(arg C.nox_format_list_arg) nativeFormatArgs {
	source := unsafe.Pointer(&arg)
	if sysVFormatListByPointer {
		source = *(*unsafe.Pointer)(source)
	}
	return *(*nativeFormatArgs)(source)
}

func nativeFormatArgsForCRT(args *nativeFormatArgs) C.nox_crt_format_list_arg {
	cursor := unsafe.Pointer(args)
	if sysVFormatListByPointer {
		return *(*C.nox_crt_format_list_arg)(unsafe.Pointer(&cursor))
	}
	return *(*C.nox_crt_format_list_arg)(cursor)
}

//export nox_vsnprintf
func nox_vsnprintf(out *C.char, capacity C.size_t, format *C.nox_format_byte, arg C.nox_format_list_arg) C.int {
	args := copyNativeFormatArgs(arg)
	return nox_format_narrow_go(out, capacity, format, &args)
}

//export nox_vsnwprintf
func nox_vsnwprintf(out *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, arg C.nox_format_list_arg) C.int {
	args := copyNativeFormatArgs(arg)
	return nox_format_wide_go(out, capacity, format, &args)
}

//export nox_vsprintf
func nox_vsprintf(out *C.char, format *C.nox_format_byte, arg C.nox_format_list_arg) C.int {
	return nox_vsnprintf(out, 0x7fffffff, format, arg)
}

//export nox_vswprintf
func nox_vswprintf(out *C.wchar2_t, format *C.nox_format_word, arg C.nox_format_list_arg) C.int {
	return nox_vsnwprintf(out, 0x3fffffff, format, arg)
}

//export nox_format_host_narrow
func nox_format_host_narrow(args *nativeFormatArgs, out *C.char, capacity C.size_t, format *C.nox_format_byte) {
	C.nox_crt_vsnprintf(out, capacity, (*C.char)(unsafe.Pointer(format)), nativeFormatArgsForCRT(args))
}
