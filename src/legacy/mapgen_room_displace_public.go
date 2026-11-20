package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

func mapgenDisplacedPosition(room, blocker *mapgenRoomGeometry) ([2]uint32, bool) {
	x, y := int32(room.grid[0]), int32(room.grid[1])
	bx, by := int32(blocker.grid[0]), int32(blocker.grid[1])
	margins := [4]int32{
		by + int32(blocker.cellDimensions[1]) - y,
		y + int32(room.cellDimensions[1]) - by,
		bx + int32(blocker.cellDimensions[0]) - x,
		x + int32(room.cellDimensions[0]) - bx,
	}
	direction, distance := -1, int32(99999)
	// The original tie order is Y+, Y-, X+, X-.
	for candidate, size := range margins {
		if size < distance {
			direction, distance = candidate, size
		}
	}
	if direction < 0 {
		return [2]uint32{}, false
	}

	axis := 1 - direction/2
	position := [2]uint32{uint32(room.position[0]), uint32(room.position[1])}
	delta := mulFloat64TowardZero(float64(distance), mapgenCellSize)
	if direction%2 != 0 {
		delta = -delta
	}
	start := float64(math.Float32frombits(position[axis]))
	position[axis] = math.Float32bits(storeFloat32TowardZero(addFloat64TowardZero(start, delta)))
	return position, true
}

func mapgenDisplaceRoom(room, blocker *mapgenRoomGeometry, move func(*mapgenRoomGeometry, [2]uint32) *mapgenRoomGeometry) bool {
	for attempt := 0; attempt < 25; attempt++ {
		position, ok := mapgenDisplacedPosition(room, blocker)
		if !ok {
			return false
		}
		blocker = move(room, position)
		if blocker == nil {
			return true
		}
	}
	return false
}

//export sub_5212B0
func sub_5212B0(rawRoom *C.uint8_t, rawBlocker *C.uint32_t) C.int {
	room := (*mapgenRoomGeometry)(unsafe.Pointer(rawRoom))
	blocker := (*mapgenRoomGeometry)(unsafe.Pointer(rawBlocker))
	return C.int(bool2int(mapgenDisplaceRoom(room, blocker, func(room *mapgenRoomGeometry, coordinates [2]uint32) *mapgenRoomGeometry {
		nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(unsafe.Pointer(room)),
			(*C.float2)(unsafe.Pointer(&coordinates[0])))
		return (*mapgenRoomGeometry)(nox_mapgenRoomAtNative_521200((*C.uint8_t)(unsafe.Pointer(room))))
	})))
}
