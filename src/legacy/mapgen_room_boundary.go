package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

func mapgenRoomFloat(bits C.uint32_t) C.float {
	return C.float(math.Float32frombits(uint32(bits)))
}

func mapgenPlaceRoomBoundary(room *mapgenRoomGeometry) {
	left, top := mapgenRoomFloat(room.bounds[0]), mapgenRoomFloat(room.bounds[1])
	right, bottom := mapgenRoomFloat(room.bounds[2]), mapgenRoomFloat(room.bounds[3])
	horizontalCount, verticalCount := C.int(room.cellDimensions[0]-1), C.int(room.cellDimensions[1]-1)

	for _, edge := range [...]struct {
		y              C.float
		leftDirection  C.int
		rightDirection C.int
	}{
		{top, 8, 9},
		{bottom, 7, 10},
	} {
		point := C.float2{field_0: left, field_4: edge.y}
		sub_526D50(edge.leftDirection)
		sub_526E60(&point.field_0)
		point.field_0 += mapgenCellSize
		sub_524500(&point, horizontalCount)

		point = C.float2{field_0: right, field_4: edge.y}
		sub_526D50(edge.rightDirection)
		sub_526E60(&point.field_0)
	}
	for _, x := range [...]C.float{left, right} {
		point := C.float2{field_0: x, field_4: top + mapgenCellSize}
		mapgenPlaceWallStrip(&point, verticalCount, 1)
	}
}

//export nox_mapgenGenerateRoomNative_524E00
func nox_mapgenGenerateRoomNative_524E00(theme, raw *C.uint8_t) {
	if theme == nil || raw == nil {
		return
	}
	ranked := (*mapgenRankedRoom)(unsafe.Pointer(raw))
	if ranked.flags&2 != 0 {
		return
	}
	surface := mapgenResolveWallFloor(C.uint32_t(sub_5244D0(C.int(ranked.decor))))
	if surface == nil {
		return
	}
	nox_mapgenGenerateWallFloorNative_524680(theme, raw, surface)
	sub_526CA0(&surface.wall_name[0])
	mapgenPlaceRoomBoundary((*mapgenRoomGeometry)(unsafe.Pointer(raw)))

	sub_526C80(0)
	mapgenVisitRoomNeighbors(raw, func(neighbor *C.uint8_t, direction C.int) {
		nox_mapgenReserveAdjacentNative_524FB0(raw, neighbor, direction)
	})
	sub_526C80(1)
	nox_mapgenGenerateDoorsNative_525510(theme, raw)
}

func mapgenVisitRoomNeighbors(raw *C.uint8_t, visit func(*C.uint8_t, C.int)) {
	links := (*mapgenRoomLinkRecord)(unsafe.Pointer(raw))
	for direction := range links.links {
		count := min(int(links.counts[direction]), len(links.links[direction]))
		for _, token := range links.links[direction][:count] {
			neighbor := mapgenResolveRoomLinkRecord(token)
			if neighbor != nil {
				visit((*C.uint8_t)(unsafe.Pointer(neighbor)), C.int(direction))
			}
		}
	}
}

//export nox_xxx_gen_524E00
func nox_xxx_gen_524E00(theme, room C.int) {
	nox_mapgenGenerateRoomNative_524E00(mapgenThemeByToken(theme), mapgenThemeByToken(room))
}

func mapgenAdjacentRoomEdge(room, neighbor *mapgenRoomGeometry, direction C.int) (start [2]C.float, axis int, far bool) {
	axis = int(direction / 2)
	cross := 1 - axis
	edge := cross
	far = direction == 1 || direction == 2
	if far {
		edge += 2
	}

	start[cross] = mapgenRoomFloat(room.bounds[edge])
	start[axis] = mapgenRoomFloat(room.bounds[axis])
	if candidate := mapgenRoomFloat(neighbor.bounds[axis]); candidate >= start[axis] {
		start[axis] = candidate
	}
	return start, axis, far
}

//export nox_mapgenReserveAdjacentNative_524FB0
func nox_mapgenReserveAdjacentNative_524FB0(raw, other *C.uint8_t, direction C.int) C.int {
	if raw == nil || other == nil {
		return 0
	}
	if direction < 0 || direction > 3 {
		return direction
	}
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	neighbor := (*mapgenRoomGeometry)(unsafe.Pointer(other))
	start, axis, far := mapgenAdjacentRoomEdge(room, neighbor, direction)
	reserved := start
	if far {
		cross := 1 - axis
		reserved[cross] = C.float(float64(reserved[cross]) - mapgenHallCellSize)
	}
	short := neighbor
	if room.cellDimensions[axis] < neighbor.cellDimensions[axis] {
		short = room
	}
	size := [2]C.float{mapgenCellSize, mapgenCellSize}
	size[axis] = mapgenRoomFloat(short.dimensions[axis])
	nox_mapgenAddOccupiedRectNative_521BC0(raw, (*C.float2)(unsafe.Pointer(&reserved)), size[0], size[1])

	middle := start
	middle[axis] = C.float(float64(middle[axis]) + mapgenHallCellSize)
	mapgenRemoveWallStrip((*C.float2)(unsafe.Pointer(&middle)), C.int(short.cellDimensions[axis]-1), axis)
	sub_5253B0(&start[0])

	end := start
	end[axis] = mapgenRoomFloat(room.bounds[axis+2])
	if candidate := mapgenRoomFloat(neighbor.bounds[axis+2]); candidate <= end[axis] {
		end[axis] = candidate
	}
	return sub_5253B0(&end[0])
}

//export sub_524FB0
func sub_524FB0(room, neighbor, direction C.int) C.int {
	return nox_mapgenReserveAdjacentNative_524FB0(mapgenThemeByToken(room), mapgenThemeByToken(neighbor), direction)
}

//export sub_5253B0
func sub_5253B0(raw *C.float) C.int {
	origin := *(*[2]C.float)(unsafe.Pointer(raw))
	var mask C.int
	for i, offset := range [...]struct {
		axis  int
		delta float64
	}{
		{1, -mapgenHallCellSize},
		{1, mapgenHallCellSize},
		{0, -mapgenHallCellSize},
		{0, mapgenHallCellSize},
	} {
		point := origin
		point[offset.axis] = C.float(float64(point[offset.axis]) + offset.delta)
		var wall unsafe.Pointer
		if sub_526DD0((*C.float2)(unsafe.Pointer(&point)), &wall) != 0 {
			mask |= 1 << i
		}
	}
	directions := [...]C.int{-1, -1, -1, 1, -1, 10, 9, 6, -1, 7, 8, 4, 0, 3, 5, 2}
	if direction := directions[mask]; direction >= 0 {
		sub_526D50(direction)
		return sub_526E60(raw)
	}
	return mask - 3
}
