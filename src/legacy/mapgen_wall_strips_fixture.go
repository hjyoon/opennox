package legacy

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_3835368, dword_5d4594_3835372;
*/
import "C"

import "unsafe"

func mapgenWallStripFixture(x, y float32, count int32, vertical bool) int32 {
	merge := C.dword_5d4594_3835368
	repeat := C.dword_5d4594_3835372
	defer func() {
		C.dword_5d4594_3835368 = merge
		C.dword_5d4594_3835372 = repeat
	}()
	C.dword_5d4594_3835368 = 0
	C.dword_5d4594_3835372 = 0
	point := C.float2{field_0: C.float(x), field_4: C.float(y)}
	if vertical {
		return int32(C.sub_524550((*C.int)(unsafe.Pointer(&point)), C.int(count)))
	}
	return int32(C.sub_524500(&point, C.int(count)))
}
