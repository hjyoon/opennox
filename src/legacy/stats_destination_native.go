package legacy

/*
#include "GAME1_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export sub_4282D0
func sub_4282D0(address *C.char, port C.int) *C.char {
	dst := memmap.PtrT[[16]byte](0x5D4594, 741316)
	src := unsafe.Slice((*byte)(unsafe.Pointer(address)), len(dst))
	n := alloc.StrLenS(src)
	copy(dst[:], src[:n])
	clear(dst[n:])
	*memmap.PtrUint32(0x5D4594, 741304) = uint32(port)
	return (*C.char)(unsafe.Pointer(dst))
}
