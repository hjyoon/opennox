package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

const mapgenHallCellSize = 32.526913

func mapgenIsHall(kind uint32) bool {
	return kind >= 2 && kind <= 5
}

//export nox_xxx_mapGenCheckRoomType_5238F0
func nox_xxx_mapGenCheckRoomType_5238F0(kind *C.int) C.int {
	return C.int(bool2int(mapgenIsHall(uint32(*kind))))
}

//export sub_523920
func sub_523920(kind C.int) C.int {
	if mapgenIsHall(uint32(kind)) {
		return kind - 2
	}
	return 0
}

//export sub_523970
func sub_523970(kind C.int) C.int {
	switch kind {
	case 2:
		return 3
	case 3:
		return 2
	case 5:
		return 1
	default:
		return 0
	}
}

//export sub_5239B0
func sub_5239B0(kind C.int) C.int {
	return sub_523960(sub_523970(kind))
}

//export nox_mapgenMakeHallStructNative_523E30
func nox_mapgenMakeHallStructNative_523E30(kind, width, depth C.int) *C.float {
	raw, _ := mapgenAllocateDecorRecord(mapgenRoomRecordSize)
	if raw == nil {
		return nil
	}
	room := (*mapgenRoomGeometry)(raw)
	room.kind = C.uint32_t(kind)
	switch kind {
	case 2, 3:
		room.cellDimensions = [2]C.int32_t{C.int32_t(max(width, 1)), C.int32_t(depth)}
	case 4, 5:
		room.cellDimensions = [2]C.int32_t{C.int32_t(depth), C.int32_t(max(width, 1))}
	}
	for i, cells := range room.cellDimensions {
		world := float32(float64(cells) * mapgenHallCellSize)
		room.dimensions[i] = C.uint32_t(math.Float32bits(world))
	}
	return (*C.float)(raw)
}

//export sub_523E30
func sub_523E30(kind, width, depth C.int) *C.float {
	return nox_mapgenMakeHallStructNative_523E30(kind, width, depth)
}

//export nox_mapgenMakeHallNative_523EC0
func nox_mapgenMakeHallNative_523EC0(raw *C.uint8_t, kind, size C.int) *C.float {
	if raw == nil {
		return nil
	}
	theme := (*[mapgenThemeWordCount]C.int)(unsafe.Pointer(raw))
	depth := nox_xxx_mapGenRandFunc2_526B00(theme[1], theme[2])
	return nox_mapgenMakeHallStructNative_523E30(kind, size, depth)
}

//export nox_xxx_mapGenMakeHall_523EC0
func nox_xxx_mapGenMakeHall_523EC0(theme, kind, size C.int) *C.float {
	return nox_mapgenMakeHallNative_523EC0(mapgenThemeByToken(theme), kind, size)
}
