package cfixture

/*
#include "GAME4_2.h"
#include "world_native_storage.h"

*/
import "C"

import "unsafe"

func DisableMapgenWallFixtureModes() func() {
	merge := C.nox_archive_mapgen_storage().merge_walls
	repeat := C.nox_archive_mapgen_storage().repeat_walls
	C.nox_archive_mapgen_storage().merge_walls = 0
	C.nox_archive_mapgen_storage().repeat_walls = 0
	return func() {
		C.nox_archive_mapgen_storage().merge_walls = merge
		C.nox_archive_mapgen_storage().repeat_walls = repeat
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
	C.nox_archive_mapgen_storage().repeat_walls = 1
}
