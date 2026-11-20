package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

func mapGenRoomOffset521B00(room, other *mapgenRoomGeometry, axis int, attach bool, roll func(int32, int32) int32) float64 {
	span := int32(room.cellDimensions[axis]) - int32(other.cellDimensions[axis])
	cells := roll(0, span)
	delta := mulFloat64TowardZero(float64(cells), mapgenCellSize)
	if attach {
		edge := math.Float32frombits(uint32(other.position[axis]))
		return addFloat64TowardZero(float64(edge), -delta)
	}
	edge := math.Float32frombits(uint32(room.position[axis]))
	return addFloat64TowardZero(delta, float64(edge))
}

func mapGenRoomOffsetNative(room, other *C.nox_mapgen_readonly_room_t, axis int, attach bool) C.float {
	value := mapGenRoomOffset521B00((*mapgenRoomGeometry)(unsafe.Pointer(room)),
		(*mapgenRoomGeometry)(unsafe.Pointer(other)), axis, attach, func(low, high int32) int32 {
			return int32(C.nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
		})
	// Both original callers spill the result to a binary32 coordinate.
	return C.float(storeFloat32TowardZero(value))
}

//export nox_mapgenRoomRandPosXNative_521B00
func nox_mapgenRoomRandPosXNative_521B00(room, child *C.nox_mapgen_readonly_room_t) C.float {
	return mapGenRoomOffsetNative(room, child, 0, false)
}

//export nox_mapgenRoomRandPosYNative_521B30
func nox_mapgenRoomRandPosYNative_521B30(room, child *C.nox_mapgen_readonly_room_t) C.float {
	return mapGenRoomOffsetNative(room, child, 1, false)
}

//export nox_mapgenRoomAttachPosXNative_521B60
func nox_mapgenRoomAttachPosXNative_521B60(room, parent *C.nox_mapgen_readonly_room_t) C.float {
	return mapGenRoomOffsetNative(room, parent, 0, true)
}

//export nox_mapgenRoomAttachPosYNative_521B90
func nox_mapgenRoomAttachPosYNative_521B90(room, parent *C.nox_mapgen_readonly_room_t) C.float {
	return mapGenRoomOffsetNative(room, parent, 1, true)
}
