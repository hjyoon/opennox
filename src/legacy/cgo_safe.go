//go:build safe

package legacy

/*
#cgo CFLAGS: -g -O0
#cgo CFLAGS: -fsanitize=address
#cgo LDFLAGS: -fsanitize=address
#cgo CFLAGS: -Dmalloc=nox_malloc
#cgo CFLAGS: -Drealloc=nox_realloc
#cgo CFLAGS: -Dcalloc=nox_calloc
#cgo CFLAGS: -Dfree=nox_free
#cgo CFLAGS: -Dmemset=nox_memset
#cgo CFLAGS: -Dmemcpy=nox_memcpy
#cgo CFLAGS: -Dmemcmp=nox_memcmp
#cgo CFLAGS: -Dstrlen=nox_strlen
#cgo CFLAGS: -Dstrcpy=nox_strcpy
#cgo CFLAGS: -Dstrcat=nox_strcat
#cgo CFLAGS: -Dstrcmp=nox_strcmp
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include "cgo_const.h"
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const cgoSafe = true

// The cgo builtins use a leaf malloc adapter. Under safe, the replacement
// allocator can grow the Go stack on a callback, so use the allocator directly.
func CString(s string) *C.char {
	p, _ := alloc.CString(s)
	return (*C.char)(unsafe.Pointer(p))
}

func CBytes(s []byte) unsafe.Pointer {
	p, _ := alloc.Malloc(uintptr(max(len(s), 1)))
	copy(unsafe.Slice((*byte)(p), len(s)), s)
	return p
}

func init() {
	memmap.SetRuntimeChecks(true)
}

func legacyCMapPointer(base, off, size uintptr) unsafe.Pointer {
	defer func() {
		if fault := recover(); fault != nil {
			fmt.Fprintln(os.Stderr, fault)
			C.abort()
		}
	}()
	return memmap.PtrSizeOff(base, off, size)
}

//export nox_malloc
func nox_malloc(size C.size_t) unsafe.Pointer {
	p, _ := alloc.Malloc(uintptr(size))
	return p
}

//export nox_realloc
func nox_realloc(ptr unsafe.Pointer, size C.size_t) unsafe.Pointer {
	return alloc.Realloc(ptr, uintptr(size))
}

//export nox_calloc
func nox_calloc(num, size C.size_t) unsafe.Pointer {
	return alloc.CallocC(uintptr(num), uintptr(size))
}

//export nox_free
func nox_free(ptr unsafe.Pointer) {
	if ptr != nil {
		alloc.FreePtr(ptr)
	}
}

//export nox_memcmp
func nox_memcmp(ptr1, ptr2 *C.nox_cgo_const_void, size C.size_t) C.int {
	return C.int(alloc.Memcmp(unsafe.Pointer(ptr1), unsafe.Pointer(ptr2), uintptr(size)))
}

//export nox_strlen
func nox_strlen(ptr *C.nox_cgo_const_char) C.size_t {
	return C.size_t(alloc.Strlen(unsafe.Pointer(ptr)))
}

//export nox_strcmp
func nox_strcmp(str1, str2 *C.nox_cgo_const_char) C.int {
	return C.int(alloc.Strcmp(unsafe.Pointer(str1), unsafe.Pointer(str2)))
}
