package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

const mapgenRoomRecordSize = 0x178

type mapgenRoomThemeDimensions struct {
	_          [32]byte
	sizeBounds [2]C.int32_t
	_          [16]byte
	extraDraw  C.uint32_t
}

func mapgenInitRoomGeometry(room *mapgenRoomGeometry, width, height int32) {
	room.kind = 1
	room.cellDimensions = [2]C.int32_t{C.int32_t(width), C.int32_t(height)}
	for i, cells := range room.cellDimensions {
		size := mulFloat64TowardZero(float64(cells), mapgenCellSize)
		room.dimensions[i] = C.uint32_t(math.Float32bits(storeFloat32TowardZero(size)))
	}
}

func mapgenPrepareRoomDimensions(theme *mapgenRoomThemeDimensions, between func(int32, int32) int32, draw func(int32, int32)) (int32, int32) {
	first, second := int32(theme.sizeBounds[0]), int32(theme.sizeBounds[1])
	minimumHeight := first - second
	width := max(5, between(first, second))
	heightBound := int32(mulFloat64TowardZero(float64(width), 0x1.333334p-2))
	height := max(minimumHeight, 5, between(width, heightBound))
	if theme.extraDraw != 0 {
		draw(0, 2)
	}
	return width, height
}

//export nox_xxx_mapGenMakeRoomStruct_521940
func nox_xxx_mapGenMakeRoomStruct_521940(width, height C.int) *C.float {
	raw := nox_mapgen_alloc_memory(1, mapgenRoomRecordSize)
	if raw == nil {
		return nil
	}
	if C.nox_mapgenLegacyPtrRegister(raw) == 0 {
		nox_mapgen_release_memory(raw)
		return nil
	}
	mapgenInitRoomGeometry((*mapgenRoomGeometry)(raw), int32(width), int32(height))
	return (*C.float)(raw)
}

//export nox_mapgenPrepareRoomNative_521990
func nox_mapgenPrepareRoomNative_521990(theme *C.uint8_t) *C.float {
	if theme == nil {
		return nil
	}
	width, height := mapgenPrepareRoomDimensions((*mapgenRoomThemeDimensions)(unsafe.Pointer(theme)),
		func(first, second int32) int32 {
			return int32(C.nox_xxx_mapGenRandFunc2_526B00(C.int(first), C.int(second)))
		}, func(first, second int32) {
			nox_xxx_mapGenRandFunc_526AC0(C.int(first), C.int(second))
		})
	return nox_xxx_mapGenMakeRoomStruct_521940(C.int(width), C.int(height))
}
