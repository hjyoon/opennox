//go:build safe && (386 || amd64 || arm || arm64)

package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "cgo_const.h"

_Static_assert(sizeof(uintptr_t) == sizeof(void*), "safe memory result width");

#define NOX_SAFE_STRINGIFY_(x) #x
#define NOX_SAFE_STRINGIFY(x) NOX_SAFE_STRINGIFY_(x)
#define NOX_SAFE_LABEL(name) NOX_SAFE_STRINGIFY(__USER_LABEL_PREFIX__) #name

#ifdef _WIN32
// The PE export list also uses the cgo name. Give the C entry the same address.
#define NOX_SAFE_ALIAS(name) __asm__(".globl " NOX_SAFE_LABEL(name) "\n" \
	".set " NOX_SAFE_LABEL(name) ", " NOX_SAFE_LABEL(name##_result) "\n")
NOX_SAFE_ALIAS(nox_memset);
NOX_SAFE_ALIAS(nox_memcpy);
NOX_SAFE_ALIAS(nox_strcpy);
NOX_SAFE_ALIAS(nox_strcat);
#undef NOX_SAFE_ALIAS
#else
uintptr_t nox_memset_result(void*, int, size_t) __asm__(NOX_SAFE_LABEL(nox_memset));
uintptr_t nox_memcpy_result(void*, nox_cgo_const_void*, size_t) __asm__(NOX_SAFE_LABEL(nox_memcpy));
uintptr_t nox_strcpy_result(char*, nox_cgo_const_char*) __asm__(NOX_SAFE_LABEL(nox_strcpy));
uintptr_t nox_strcat_result(char*, nox_cgo_const_char*) __asm__(NOX_SAFE_LABEL(nox_strcat));
#endif

#undef NOX_SAFE_LABEL
#undef NOX_SAFE_STRINGIFY
#undef NOX_SAFE_STRINGIFY_
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Each C caller already owns or borrows its destination. Return that same
// address in the native result register without exporting a new Go pointer.
// The C declarations keep their pointer returns.

//export nox_memset_result
func nox_memset_result(dst unsafe.Pointer, value C.int, size C.size_t) C.uintptr_t {
	return C.uintptr_t(uintptr(alloc.Memset(dst, byte(value), uintptr(size))))
}

//export nox_memcpy_result
func nox_memcpy_result(dst unsafe.Pointer, src *C.nox_cgo_const_void, size C.size_t) C.uintptr_t {
	return C.uintptr_t(uintptr(alloc.Memcpy(dst, unsafe.Pointer(src), uintptr(size))))
}

//export nox_strcpy_result
func nox_strcpy_result(dst *C.char, src *C.nox_cgo_const_char) C.uintptr_t {
	return C.uintptr_t(uintptr(alloc.Strcpy(unsafe.Pointer(dst), unsafe.Pointer(src))))
}

//export nox_strcat_result
func nox_strcat_result(dst *C.char, src *C.nox_cgo_const_char) C.uintptr_t {
	return C.uintptr_t(uintptr(alloc.Strcat(unsafe.Pointer(dst), unsafe.Pointer(src))))
}
