package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func mapgenScanAdjacentPortals(room *mapgenRankedRoom, direction int,
	lookup func([2]int32) *mapgenRankedRoom,
	place func(*mapgenRankedRoom, int32, bool),
) {
	axis := direction / 2
	normal := 1 - axis
	length := int32(room.cellDimensions[axis])
	cell := [2]int32{int32(room.grid[0]), int32(room.grid[1])}
	if direction == 1 || direction == 2 {
		cell[normal] += int32(room.cellDimensions[normal])
	} else {
		cell[normal]--
	}
	var previous *mapgenRankedRoom
	var first, count int32
	for i := int32(0); i < length; i++ {
		current := lookup(cell)
		if current != nil && current == previous && current.flags&2 != 0 {
			count++
		}
		if current != previous || i == length-1 {
			if previous != nil && count >= 3 {
				place(previous, (first+i)/2, count >= 4)
			}
			first, count = i, 1
			previous = current
			if current != nil && current.kind == 1 && (direction == 1 || direction == 3) {
				previous = nil
			}
		}
		cell[axis]++
	}
}

const mapgenPortalHalfCell = 16.263456

func mapgenReservePortalCells(room, neighbor *mapgenRankedRoom, point [2]C.float, normal int, insideNegative bool) {
	const size = C.float(mapgenHallCellSize)
	if insideNegative {
		point[normal] = C.float(float64(point[normal]) - mapgenHallCellSize)
	}
	nox_mapgenAddOccupiedRectNative_521BC0((*C.uint8_t)(unsafe.Pointer(room)), (*C.float2)(unsafe.Pointer(&point)), size, size)
	if mapgenIsHall(uint32(neighbor.kind)) {
		side := -mapgenHallCellSize
		if insideNegative {
			side = mapgenHallCellSize
		}
		point[normal] = C.float(float64(point[normal]) + side)
		nox_mapgenAddOccupiedRectNative_521BC0((*C.uint8_t)(unsafe.Pointer(neighbor)), (*C.float2)(unsafe.Pointer(&point)), size, size)
	}
}

func mapgenPlaceAdjacentPortal(room, neighbor *mapgenRankedRoom, direction int, midpoint int32, double bool) {
	axis := direction / 2
	normal := 1 - axis
	insideNegative := direction == 1 || direction == 2
	bounds := *(*[4]C.float)(unsafe.Pointer(&room.bounds))
	alignedBase := float64(bounds[axis])
	aligned := float64(float64(midpoint)*mapgenHallCellSize) + alignedBase
	if runtime.GOARCH == "arm64" {
		aligned = math.FMA(float64(midpoint), mapgenHallCellSize, alignedBase)
	}
	var point [2]C.float
	point[axis] = C.float(aligned)
	point[normal] = bounds[normal]
	if insideNegative {
		point[normal] = bounds[normal+2]
	}
	position := (*C.float2)(unsafe.Pointer(&point))
	sub_527030(position)
	point[axis] = C.float(float64(point[axis]) - mapgenPortalHalfCell)
	baseName := "ArchedDoor"
	if double {
		baseName = "ArchedHalfDoor"
	}
	orientations := [...][2]C.int{{5, 3}, {7, 1}}[axis]
	nox_xxx_mapGenGetObjID_527940(internCStr(baseName))
	mapgenPlaceDoorObject(position, orientations[0])
	mapgenReservePortalCells(room, neighbor, point, normal, insideNegative)

	center := float64(point[axis]) + mapgenPortalHalfCell
	if double {
		point[axis] = C.float(center + mapgenHallCellSize)
		sub_527030(position)
		point[axis] = C.float(float64(point[axis]) + mapgenPortalHalfCell)
		nox_xxx_mapGenGetObjID_527940(internCStr(baseName))
		mapgenPlaceDoorObject(position, orientations[1])
		point[axis] = C.float(float64(point[axis]) - mapgenHallCellSize)
		mapgenReservePortalCells(room, neighbor, point, normal, insideNegative)
		center = float64(point[axis])
	}

	near, far := point, point
	near[axis], far[axis] = C.float(center), C.float(center)
	near[normal] = C.float(float64(point[normal]) - mapgenHallCellSize)
	far[normal] = C.float(float64(point[normal]) + mapgenHallCellSize)
	lower, upper := (*C.float2)(unsafe.Pointer(&near)), (*C.float2)(unsafe.Pointer(&far))
	sub_522C80(lower)
	sub_522C80(upper)
	sub_51D3F0(lower, upper)
	sub_51D3F0(upper, lower)

	current, adjacent := upper, lower
	if insideNegative {
		current, adjacent = lower, upper
	}
	rawRoom, rawNeighbor := (*C.uint8_t)(unsafe.Pointer(room)), (*C.uint8_t)(unsafe.Pointer(neighbor))
	nox_mapgenAddRoomPointNative_522CA0(rawRoom, current)
	if neighbor.kind == 1 {
		nox_mapgenAddRoomPointNative_522CA0(rawNeighbor, adjacent)
	}
	nox_mapgenRoomLinkNative_521900(rawRoom, rawNeighbor, C.int(direction))
	nox_mapgenRoomLinkNative_521900(rawNeighbor, rawRoom, sub_523960(C.int(direction)))
}

//export nox_xxx_mapgen_Doors_4D4790
func nox_xxx_mapgen_Doors_4D4790() *C.float {
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		if room.kind != 1 {
			continue
		}
		nox_xxx_mapGenSetFlags_5235F0(-100)
		links := (*mapgenRoomLinkRecord)(unsafe.Pointer(room))
		for direction := range links.counts {
			if links.counts[direction] != 0 ||
				nox_xxx_mapGenRandFunc_526AC0(1, 100) > C.int(memmap.Int32(0x5D4594, 1549848)) {
				continue
			}
			mapgenScanAdjacentPortals(room, direction, func(cell [2]int32) *mapgenRankedRoom {
				position := C.int2{field_0: C.int(cell[0]), field_4: C.int(cell[1])}
				return (*mapgenRankedRoom)(nox_mapgenRoomAtCellNative_521290(&position))
			}, func(neighbor *mapgenRankedRoom, midpoint int32, double bool) {
				mapgenPlaceAdjacentPortal(room, neighbor, direction, midpoint, double)
			})
		}
	}
	return nil
}
