package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

func mapgenPlaceDoorObject(position *C.float2, direction C.int) {
	if object := nox_xxx_mapGenPlaceObj_5279B0(position); object != nil {
		nox_mapgenOrientObjNative_527C60(object, direction)
	}
}

func mapgenPlaceRoomDoor(raw *C.uint8_t, point *C.float2, span C.int, axis int, double bool) C.int {
	if raw == nil {
		return 0
	}
	room := (*mapgenRankedRoom)(unsafe.Pointer(raw))
	definition := mapgenResolveDecorDefinition(room.decor)
	if definition == nil {
		return 0
	}
	name := &definition.door_name[0]
	if double {
		name = &definition.double_door_name[0]
	}
	if *name == 0 {
		return 0
	}

	mapgenPlaceWallStrip(point, span+1, axis)
	coordinates := *(*[2]C.float)(unsafe.Pointer(point))
	coordinates[axis] = C.float(float64(span/2)*mapgenHallCellSize + float64(coordinates[axis]))
	position := (*C.float2)(unsafe.Pointer(&coordinates))
	sub_527030(position)
	if double {
		coordinates[axis] += mapgenCellSize
		sub_527030(position)
		coordinates[axis] -= 48.790367
	} else {
		coordinates[axis] -= 16.263456
	}

	directions := [...][2]C.int{{5, 3}, {7, 1}}[axis]
	nox_xxx_mapGenGetObjID_527940(name)
	mapgenPlaceDoorObject(position, directions[0])
	if double {
		coordinates[axis] += 65.053825
		mapgenPlaceDoorObject(position, directions[1])
	}
	return 1
}

func mapgenPlaceDoorBetweenRooms(raw, other *C.uint8_t, direction C.int) {
	if raw == nil || other == nil || direction < 0 || direction > 3 {
		return
	}
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	neighbor := (*mapgenRoomGeometry)(unsafe.Pointer(other))
	start, axis, _ := mapgenAdjacentRoomEdge(room, neighbor, direction)
	span := C.int(min(room.cellDimensions[axis], neighbor.cellDimensions[axis]))
	if span < 2 {
		return
	}
	point := (*C.float2)(unsafe.Pointer(&start))
	if span == 2 || mapgenPlaceRoomDoor(raw, point, span, axis, true) == 0 {
		mapgenPlaceRoomDoor(raw, point, span, axis, false)
	}
}

//export nox_mapgenGenerateDoorsNative_525510
func nox_mapgenGenerateDoorsNative_525510(theme, room *C.uint8_t) {
	if theme == nil || room == nil {
		return
	}
	mapgenVisitRoomNeighbors(room, func(neighbor *C.uint8_t, direction C.int) {
		mapgenPlaceDoorBetweenRooms(room, neighbor, direction)
	})
}

//export nox_xxx_mapgen_525510
func nox_xxx_mapgen_525510(theme, room C.int) {
	nox_mapgenGenerateDoorsNative_525510(mapgenThemeByToken(theme), mapgenThemeByToken(room))
}

//export nox_xxx_mapgen_525570
func nox_xxx_mapgen_525570(_ C.int, room, neighbor, direction C.int) {
	mapgenPlaceDoorBetweenRooms(mapgenThemeByToken(room), mapgenThemeByToken(neighbor), direction)
}

//export nox_xxx_mapgen_525690
func nox_xxx_mapgen_525690(room C.int, point *C.float2, span C.int) C.int {
	return mapgenPlaceRoomDoor(mapgenThemeByToken(room), point, span, 0, false)
}

//export nox_xxx_mapgen_525740
func nox_xxx_mapgen_525740(room C.int, point *C.float2, span C.int) C.int {
	return mapgenPlaceRoomDoor(mapgenThemeByToken(room), point, span, 0, true)
}

//export nox_xxx_mapgen_525830
func nox_xxx_mapgen_525830(room C.int, point *C.float2, span C.int) C.int {
	return mapgenPlaceRoomDoor(mapgenThemeByToken(room), point, span, 1, false)
}

//export nox_xxx_mapgen_5258E0
func nox_xxx_mapgen_5258E0(room C.int, point *C.float2, span C.int) C.int {
	return mapgenPlaceRoomDoor(mapgenThemeByToken(room), point, span, 1, true)
}
