package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

//export nox_mapgenAddRoomPointNative_522CA0
func nox_mapgenAddRoomPointNative_522CA0(raw *C.uint8_t, source *C.nox_mapgen_readonly_point_t) C.char {
	if raw == nil || source == nil {
		return 0
	}
	room := (*mapgenRankedRoom)(unsafe.Pointer(raw))
	point := (*C.float2)(unsafe.Pointer(source))
	count := int(room.pointCount)
	if count >= len(room.points) {
		return C.char(count)
	}
	for _, old := range room.points[:count] {
		if sub_524660(old.field_0, point.field_0) != 0 && sub_524660(old.field_4, point.field_4) != 0 {
			return 1
		}
	}
	room.points[count] = *point
	room.pointCount++
	return C.char(room.pointCount)
}

//export sub_522CA0
func sub_522CA0(room C.int, point *C.float) C.char {
	return nox_mapgenAddRoomPointNative_522CA0(mapgenThemeByToken(room),
		(*C.nox_mapgen_readonly_point_t)(unsafe.Pointer(point)))
}

func mapgenBidirectionalRoomWaypoints(theme *C.uint8_t) bool {
	if theme == nil {
		return true
	}
	words := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(theme))
	return words[15] == 0
}

func mapgenConnectRoomWaypointPair(theme *C.uint8_t, from, to *C.float2) {
	sub_522C80(from)
	sub_522C80(to)
	sub_51D3F0(from, to)
	if mapgenBidirectionalRoomWaypoints(theme) {
		sub_51D3F0(to, from)
	}
}

func mapgenConnectRoomHallWaypoints(theme, raw *C.uint8_t) {
	mapgenVisitRoomNeighbors(raw, func(neighbor *C.uint8_t, side C.int) {
		kind := (*mapgenRankedRoom)(unsafe.Pointer(neighbor)).kind
		if !mapgenIsHall(uint32(kind)) || sub_523920(C.int(kind)) != side {
			return
		}
		var from, to C.float2
		nox_mapgenConnectionPointFarNative_523CB0(neighbor, &from)
		nox_mapgenAddRoomPointNative_522CA0(raw, (*C.nox_mapgen_readonly_point_t)(unsafe.Pointer(&from)))
		nox_mapgenConnectionPointCenterNative_523D30(neighbor, &to)
		mapgenConnectRoomWaypointPair(theme, &from, &to)
	})
}

func mapgenConnectHallWaypoints(theme, raw *C.uint8_t) {
	var from C.float2
	nox_mapgenConnectionPointCenterNative_523D30(raw, &from)
	mapgenVisitRoomNeighbors(raw, func(neighbor *C.uint8_t, side C.int) {
		if sub_521AA0((*C.uint32_t)(unsafe.Pointer(raw)), side) != 0 {
			return
		}
		var to C.float2
		kind := (*mapgenRankedRoom)(unsafe.Pointer(neighbor)).kind
		if mapgenIsHall(uint32(kind)) {
			nox_mapgenConnectionPointCenterNative_523D30(neighbor, &to)
		} else {
			nox_mapgenConnectionPointNearNative_523C30(raw, &to)
			nox_mapgenAddRoomPointNative_522CA0(neighbor, (*C.nox_mapgen_readonly_point_t)(unsafe.Pointer(&to)))
		}
		mapgenConnectRoomWaypointPair(theme, &from, &to)
	})
}

//export nox_mapgenConnectRoomLinesNative_522D30
func nox_mapgenConnectRoomLinesNative_522D30(theme *C.uint8_t) *C.float {
	sub_51D0F0(-128)
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		nox_xxx_mapGenSetFlags_5235F0(-100)
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		raw := (*C.uint8_t)(unsafe.Pointer(room))
		if room.kind == 1 {
			mapgenConnectRoomHallWaypoints(theme, raw)
		} else {
			mapgenConnectHallWaypoints(theme, raw)
		}
	}
	return nil
}

//export sub_522D30
func sub_522D30(theme C.int) *C.float {
	return nox_mapgenConnectRoomLinesNative_522D30(mapgenThemeByToken(theme))
}

//export nox_mapgenConnectRoomPointsNative_522F40
func nox_mapgenConnectRoomPointsNative_522F40(theme *C.uint8_t) *C.uint8_t {
	if theme == nil || !mapgenBidirectionalRoomWaypoints(theme) {
		return theme
	}
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		if room.kind != 1 {
			continue
		}
		nox_xxx_mapGenSetFlags_5235F0(-100)
		points := room.points[:min(int(room.pointCount), len(room.points))]
		for first := range points {
			for second := range points {
				if first != second {
					sub_51D3F0(&points[first], &points[second])
				}
			}
		}
	}
	return nil
}

//export nox_xxx_mapGenTryNextRoom_522F40
func nox_xxx_mapGenTryNextRoom_522F40(theme *C.uint32_t) *C.uint8_t {
	raw := (*C.uint8_t)(resolveMapgenPointer(uintptr(unsafe.Pointer(theme))))
	return nox_mapgenConnectRoomPointsNative_522F40(raw)
}
