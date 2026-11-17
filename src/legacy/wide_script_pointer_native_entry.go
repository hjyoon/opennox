package legacy

/*
#include "defs.h"
extern unsigned int dword_5d4594_1599628;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_512E80
func sub_512E80(value *C.wchar2_t) C.int {
	count := uint32(C.dword_5d4594_1599628)
	result := count
	if int32(count) < 1024 {
		offset := uint32(26664) + 4*count
		slot := memmap.PtrPtr(0x973F18, uintptr(offset))
		if unsafe.Sizeof(uintptr(0)) == 4 {
			// Packed C slots may contain uninitialized bytes. Store the native
			// C pointer without shading the old word through a Go barrier.
			*(*uintptr)(unsafe.Pointer(slot)) = uintptr(unsafe.Pointer(value))
		} else {
			// On 64-bit builds this is an initialized Go pointer side slot.
			*slot = unsafe.Pointer(value)
		}
		result = count + 1
		C.dword_5d4594_1599628 = C.uint(result)
	}
	return C.int(int32(result - 1))
}
