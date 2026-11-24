package legacy

/*
#include "GAME1_1.h"
#include "GAME2_3.h"

extern uint32_t dword_5d4594_1305788;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export sub_49FDB0
func sub_49FDB0(region C.int) {
	if C.dword_5d4594_1305788 != 0 {
		return
	}
	const base = 0x587000
	const vertexBase = 165104
	const recordStart, recordEnd, recordSize = 165744, 166016, 16
	for record := uintptr(recordStart); record < recordEnd; record += recordSize {
		for i := range int(memmap.Int8(base, record)) {
			index := int(memmap.Int8(base, record+1+uintptr(i)))
			vertex := uintptr(vertexBase + 8*index)
			C.sub_420DA0(C.float(memmap.Float32(base, vertex)), C.float(memmap.Float32(base, vertex+4)))
		}
		name := GoStringP(*memmap.PtrPtr(base, record+12))
		var polygon [140]byte
		alloc.StrCopyZero(polygon[4:104], name)
		C.sub_4211D0(unsafe.Pointer(&polygon[0]))
		C.sub_4214D0()
	}
	C.dword_5d4594_1305788 = 1
}

//export sub_49FF20
func sub_49FF20() *C.uint32_t {
	if C.dword_5d4594_1305788 == 0 {
		return nil
	}
	result := C.sub_421B10()
	C.dword_5d4594_1305788 = 0
	return result
}
