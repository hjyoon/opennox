package legacy

/*
#include "GAME4_2.h"

static inline int nox_mapgen_call_prefab_connector(nox_mapgen_prefab_connect_cb_5266f0 callback,
	uint8_t* prefab, int direction, uint8_t* candidate) {
	return callback(prefab, direction, candidate);
}
*/
import "C"

import (
	"cmp"
	"slices"
	"unsafe"
)

type mapgenPrefabCandidate struct {
	room     *mapgenRankedRoom
	distance uint64
}

func mapgenRoomOnPrefabSide(dx, dy int64, direction C.int) bool {
	switch direction {
	case mapgenPrefabNorth5262F0:
		return dy < 0
	case mapgenPrefabSouth5262F0:
		return dy > 0
	case mapgenPrefabEast5262F0:
		return dx > 0
	case mapgenPrefabWest5262F0:
		return dx < 0
	default:
		return false
	}
}

func mapgenClosestPrefabRooms(marker mapgenPrefabExit5262F0, direction C.int) *mapgenRankedRoom {
	var closest [6]mapgenPrefabCandidate
	count := 0
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		if mapgenIsHall(uint32(room.kind)) {
			continue
		}
		dx := int64(int32(room.grid[0])) + int64(room.cellDimensions[0]/2) - int64(marker.x)
		dy := int64(int32(room.grid[1])) + int64(room.cellDimensions[1]/2) - int64(marker.y)
		if !mapgenRoomOnPrefabSide(dx, dy, direction) {
			continue
		}
		candidate := mapgenPrefabCandidate{room: room, distance: uint64(dx*dx + dy*dy)}
		if count < len(closest) {
			closest[count] = candidate
			count++
			continue
		}
		farthest := 0
		for i := 1; i < count; i++ {
			if closest[i].distance > closest[farthest].distance {
				farthest = i
			}
		}
		if candidate.distance < closest[farthest].distance {
			closest[farthest] = candidate
		}
	}
	slices.SortStableFunc(closest[:count], func(a, b mapgenPrefabCandidate) int {
		return cmp.Compare(a.distance, b.distance)
	})
	for i, candidate := range closest[:count] {
		var next uint32
		if i+1 < count {
			next = mapgenRankedRoomToken(closest[i+1].room)
		}
		candidate.room.prefabNext = next
	}
	return closest[0].room
}

//export nox_mapgenClosestRoomsNative_526550
func nox_mapgenClosestRoomsNative_526550(raw *C.uint8_t, direction C.int) *C.uint8_t {
	if raw == nil || direction < 0 || direction >= mapgenPrefabDirections5262F0 {
		return nil
	}
	prefab := (*mapgenThemePrefabRecord)(unsafe.Pointer(raw))
	return (*C.uint8_t)(unsafe.Pointer(mapgenClosestPrefabRooms(prefab.markers[direction], direction)))
}

//export sub_526550
func sub_526550(prefab, direction C.int) C.int {
	first := nox_mapgenClosestRoomsNative_526550(mapgenThemeByToken(prefab), direction)
	return C.int(mapgenRankedRoomToken((*mapgenRankedRoom)(unsafe.Pointer(first))))
}

func mapgenReplacePrefabRoom(prefab *mapgenThemePrefabRecord) bool {
	old := mapgenRankedRoomByToken(prefab.room)
	if old == nil {
		return false
	}
	inset := float64(int32(prefab.maxMarkerCount)) * 65.053825
	position := C.float2{
		field_0: C.float(float64(mapgenRoomFloat(old.position[0])) + inset),
		field_4: C.float(float64(mapgenRoomFloat(old.position[1])) + inset),
	}
	sub_521760(C.int(prefab.room))
	sub_521A10(unsafe.Pointer(old))
	prefab.room = 0

	raw := unsafe.Pointer(nox_xxx_mapGenMakeRoomStruct_521940(
		mapgenPrefabRoomCells(prefab.width, 0), mapgenPrefabRoomCells(prefab.height, 0)))
	if raw == nil {
		return false
	}
	room := (*mapgenRankedRoom)(raw)
	prefab.room = mapgenRankedRoomToken(room)
	nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(raw), &position)
	nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(raw))
	room.flags |= 2
	return true
}

func mapgenConnectPrefabs(theme *C.uint8_t, connect func(*mapgenThemePrefabRecord, C.int, *mapgenRankedRoom) bool) bool {
	words := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(theme))
	for prefab := mapgenResolveThemePrefab(words[20]); prefab != nil; prefab = mapgenResolveThemePrefab(C.uint32_t(prefab.next)) {
		if prefab.placed == 0 {
			continue
		}
		if !mapgenReplacePrefabRoom(prefab) {
			return false
		}
		for direction := range prefab.markers {
			marker := prefab.markers[direction]
			if marker.present == 0 {
				continue
			}
			side := C.int(direction)
			candidate := mapgenClosestPrefabRooms(marker, side)
			for candidate != nil && !connect(prefab, side, candidate) {
				candidate = mapgenRankedRoomByToken(candidate.prefabNext)
			}
			if candidate == nil {
				return false
			}
		}
	}
	return true
}

//export nox_mapgenConnectPrefabsWithCallback_5266F0
func nox_mapgenConnectPrefabsWithCallback_5266F0(theme *C.uint8_t, callback C.nox_mapgen_prefab_connect_cb_5266f0) C.int {
	if theme == nil || callback == nil {
		return 0
	}
	return C.int(bool2int(mapgenConnectPrefabs(theme, func(prefab *mapgenThemePrefabRecord, side C.int, candidate *mapgenRankedRoom) bool {
		return C.nox_mapgen_call_prefab_connector(callback, (*C.uint8_t)(unsafe.Pointer(prefab)), side,
			(*C.uint8_t)(unsafe.Pointer(candidate))) != 0
	})))
}

//export nox_xxx_mapGen_InPrefab2_5266F0
func nox_xxx_mapGen_InPrefab2_5266F0(theme *C.uint8_t) C.int {
	if theme == nil {
		return 0
	}
	return C.int(bool2int(mapgenConnectPrefabs(theme, func(prefab *mapgenThemePrefabRecord, side C.int, candidate *mapgenRankedRoom) bool {
		return sub_54B2D0((*C.uint8_t)(unsafe.Pointer(prefab)), side, (*C.uint8_t)(unsafe.Pointer(candidate))) != 0
	})))
}
