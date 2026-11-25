package cfixture

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_3835368, dword_5d4594_3835372;
*/
import "C"

import "unsafe"

func DisableMapgenWallFixtureModes() func() {
	merge := C.dword_5d4594_3835368
	repeat := C.dword_5d4594_3835372
	C.dword_5d4594_3835368 = 0
	C.dword_5d4594_3835372 = 0
	return func() {
		C.dword_5d4594_3835368 = merge
		C.dword_5d4594_3835372 = repeat
	}
}

func MapgenWallStripFixture(x, y float32, count int32, vertical bool) int32 {
	point := C.float2{field_0: C.float(x), field_4: C.float(y)}
	if vertical {
		return int32(C.sub_524550((*C.int)(unsafe.Pointer(&point)), C.int(count)))
	}
	return int32(C.sub_524500(&point, C.int(count)))
}

func MapgenWallJunctionFixture(x, y float32) int32 {
	point := C.float2{field_0: C.float(x), field_4: C.float(y)}
	return int32(C.sub_5253B0(&point.field_0))
}

func MapgenWallPlacementFixture(x, y float32) int32 {
	point := C.float2{field_0: C.float(x), field_4: C.float(y)}
	return int32(C.sub_526E60(&point.field_0))
}

func EnableMapgenWallFixtureRepeat() {
	C.dword_5d4594_3835372 = 1
}
