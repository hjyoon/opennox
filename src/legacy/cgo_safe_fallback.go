//go:build safe && !386 && !amd64 && !arm && !arm64

package legacy

/*
#include <stddef.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_memset_go
func nox_memset_go(dst unsafe.Pointer, value C.int, size C.size_t) {
	alloc.Memset(dst, byte(value), uintptr(size))
}

//export nox_memcpy_go
func nox_memcpy_go(dst, src unsafe.Pointer, size C.size_t) {
	alloc.Memcpy(dst, src, uintptr(size))
}

//export nox_strcpy_go
func nox_strcpy_go(dst, src *C.char) {
	alloc.Strcpy(unsafe.Pointer(dst), unsafe.Pointer(src))
}

//export nox_strcat_go
func nox_strcat_go(dst, src *C.char) {
	alloc.Strcat(unsafe.Pointer(dst), unsafe.Pointer(src))
}
