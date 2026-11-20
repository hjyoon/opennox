package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

const (
	mapgenCellSize         = 0x1.04371ep+5
	mapgenWorldToCellScale = 0x1.f7b4bap-6
)

func mapgenGridCoordinate(value float32) int64 {
	bias := 0.5
	if value < 0 {
		bias = -0.5
	}
	scaled := addFloat64TowardZero(float64(value)*mapgenWorldToCellScale, bias)
	if !(scaled >= -0x1p63 && scaled < 0x1p63) {
		return math.MinInt64
	}
	return int64(scaled)
}

//export nox_xxx_mapGenRoundFloatToPtr_520DF0
func nox_xxx_mapGenRoundFloatToPtr_520DF0(pos *C.float2, destination *C.uint32_t) C.longlong {
	cell := unsafe.Slice(destination, 2)
	cell[0] = C.uint32_t(mapgenGridCoordinate(float32(pos.field_0)))
	y := mapgenGridCoordinate(float32(pos.field_4))
	cell[1] = C.uint32_t(y)
	return C.longlong(y)
}

func mapgenRoomBoundBits(dimension, position C.uint32_t) C.uint32_t {
	size := float64(math.Float32frombits(uint32(dimension)))
	coord := float64(math.Float32frombits(uint32(position)))
	return C.uint32_t(math.Float32bits(storeFloat32TowardZero(addFloat64TowardZero(size, coord))))
}

type mapgenRoomGeometry struct {
	kind           C.uint32_t
	grid           [2]C.uint32_t
	cellDimensions [2]C.int32_t
	position       [2]C.uint32_t
	dimensions     [2]C.uint32_t
	bounds         [4]C.uint32_t
}

func updateMapgenRoomRect(room *mapgenRoomGeometry) {
	maximumX := mapgenRoomBoundBits(room.dimensions[0], room.position[0])
	room.bounds[0], room.bounds[1] = room.position[0], room.position[1]
	room.bounds[2] = maximumX
	room.bounds[3] = mapgenRoomBoundBits(room.dimensions[1], room.position[1])
}

//export nox_xxx_mapGenUpdateRoomRect_521850
func nox_xxx_mapGenUpdateRoomRect_521850(room *C.uint8_t) *C.uint8_t {
	updateMapgenRoomRect((*mapgenRoomGeometry)(unsafe.Pointer(room)))
	return room
}

//export nox_xxx_mapGenSetRoomPos_521880
func nox_xxx_mapGenSetRoomPos_521880(room *C.uint32_t, pos *C.float2) *C.uint32_t {
	record := (*mapgenRoomGeometry)(unsafe.Pointer(room))
	coordinates := (*[2]C.uint32_t)(unsafe.Pointer(pos))
	record.position[0] = coordinates[0]
	record.position[1] = coordinates[1]
	nox_xxx_mapGenRoundFloatToPtr_520DF0(pos, &record.grid[0])
	updateMapgenRoomRect(record)
	return room
}

//export sub_521AA0
func sub_521AA0(room *C.uint32_t, direction C.int) C.int {
	switch *room {
	case 1:
		return 0
	case 2:
		return C.int(bool2int(direction == 1))
	case 3:
		return C.int(bool2int(direction == 0))
	case 4:
		return C.int(bool2int(direction == 3))
	case 5:
		return C.int(bool2int(direction == 2))
	default:
		// Compare the incoming 32-bit address at the room's native width.
		return C.int(bool2int(uintptr(uint32(direction)) == uintptr(unsafe.Pointer(room))))
	}
}
