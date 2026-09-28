package legacy

/*
#include <stdint.h>

#include "GAME4.h"
#include "GAME4_3.h"
#include "common/fs/nox_fs.h"
#include "map_object_list_5048a0.h"
#include "map_temp_lists_503f40.h"
#include "server__script__file.h"

extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_1599644;
extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_3835396;
extern uint32_t dword_5d4594_3835392;
extern uint32_t dword_5d4594_3835312;

int32_t nox_script_readWriteZzz_541670(char* path, char* path2, char* dst);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type mapgenPlacePoint503B30 struct {
	x float32
	y float32
}

type mapgenPlaceGeometry503B30 struct {
	bounds   [4]int32
	delta    [2]int32
	wallSize [2]uint32
	wallSpan [2]float32
}

type mapgenPlaceDeps503B30 struct {
	fixCoords              func(mapgenPlacePoint503B30) (mapgenPlacePoint503B30, bool)
	placementReady         func() bool
	loadSelected           func() bool
	geometry               func(mapgenPlacePoint503B30, mapgenPlacePoint503B30) (mapgenPlaceGeometry503B30, bool)
	storeGeometry          func(mapgenPlaceGeometry503B30)
	placeTiles             func([2]int32) bool
	placeWalls             func([2]int32) bool
	placeWaypoints         func([2]int32) bool
	placeObjects           func([2]int32) bool
	prepareWaypoints       func()
	markPendingWaypoints   func()
	fixupPendingObjects    func([4]int32)
	placeGroups            func([2]int32) bool
	clearWaypointTempIDs   func()
	clearObjectScriptIDs   func()
	finalizeWaypoints      func()
	finalizePendingObjects func()
	markPlaced             func()
	hasScriptPayload       func() bool
	finishScriptPayload    func([2]int32)
	advancePlacementIndex  func()
}

// mapgenPlaceWithDeps503B30 preserves GAME.EXE 00503B30's partial-commit
// order. A failed placement deliberately leaves every successful earlier
// stage applied, just as the original nested call chain did.
func mapgenPlaceWithDeps503B30(at mapgenPlacePoint503B30, deps mapgenPlaceDeps503B30) bool {
	fixed, ok := deps.fixCoords(at)
	if !ok {
		return false
	}
	if !deps.placementReady() && !deps.loadSelected() {
		return false
	}
	geometry, ok := deps.geometry(at, fixed)
	if !ok {
		return false
	}
	deps.storeGeometry(geometry)
	if !deps.placeTiles(geometry.delta) {
		return false
	}
	if !deps.placeWalls(geometry.delta) {
		return false
	}
	if !deps.placeWaypoints(geometry.delta) {
		return false
	}
	if !deps.placeObjects(geometry.delta) {
		return false
	}

	deps.prepareWaypoints()
	deps.markPendingWaypoints()
	deps.fixupPendingObjects(geometry.bounds)
	if !deps.placeGroups(geometry.delta) {
		return false
	}
	deps.clearWaypointTempIDs()
	deps.clearObjectScriptIDs()
	deps.finalizeWaypoints()
	deps.finalizePendingObjects()
	deps.markPlaced()
	if deps.hasScriptPayload() {
		deps.finishScriptPayload(geometry.delta)
	}
	deps.advancePlacementIndex()
	return true
}

func mapgenPlaceRuntimeGeometry503B30(
	at mapgenPlacePoint503B30,
	fixed mapgenPlacePoint503B30,
) (mapgenPlaceGeometry503B30, bool) {
	selected := uint32(C.dword_5d4594_3835396)
	records := unsafe.Pointer(C.dword_5d4594_1599576)
	if records == nil || selected >= uint32(C.dword_5d4594_1599596) {
		return mapgenPlaceGeometry503B30{}, false
	}
	record := unsafe.Add(records, uintptr(selected)*76)
	var bounds [4]int32
	var delta [2]int32
	ok := C.nox_mapgenPlaceGeometryNative_503B30(
		C.float(at.x), C.float(at.y), C.float(fixed.x), C.float(fixed.y), record,
		C.int32_t(*memmap.PtrInt32(0x5D4594, 1599508)),
		C.int32_t(*memmap.PtrInt32(0x5D4594, 1599512)),
		(*C.int32_t)(unsafe.Pointer(&bounds[0])),
		(*C.int32_t)(unsafe.Pointer(&delta[0])),
	) != 0
	if !ok {
		return mapgenPlaceGeometry503B30{}, false
	}
	wallSize := [2]uint32{
		*memmap.PtrUint32(0x5D4594, 739980),
		*memmap.PtrUint32(0x5D4594, 739984),
	}
	return mapgenPlaceGeometry503B30{
		bounds:   bounds,
		delta:    delta,
		wallSize: wallSize,
		wallSpan: [2]float32{
			float32(C.nox_mapgenWallSpan_503B30(C.uint32_t(wallSize[0]))),
			float32(C.nox_mapgenWallSpan_503B30(C.uint32_t(wallSize[1]))),
		},
	}, true
}

func mapgenFinishScriptPayload503B30(delta [2]int32) {
	*memmap.PtrUint32(0x973F18, 35880)++
	ordinal := C.int32_t(C.dword_5d4594_3835312)
	C.sub_542BF0(ordinal, C.int32_t(delta[0]), C.int32_t(delta[1]))

	offset := [2]C.int{C.int(delta[0]), C.int(delta[1])}
	current := (*C.char)(memmap.PtrOff(0x973F18, 30760))
	previous := (*C.char)(memmap.PtrOff(0x973F18, 36008))
	merged := (*C.char)(memmap.PtrOff(0x973F18, 38056))
	C.sub_543110(current, &offset[0])
	if *memmap.PtrUint32(0x5D4594, 1599580) != 0 {
		C.nox_fs_remove(previous)
		C.nox_fs_move(merged, previous)
		C.nox_script_readWriteZzz_541670(previous, current, merged)
	} else {
		*memmap.PtrUint32(0x5D4594, 1599580) = 1
		C.nox_fs_move(current, merged)
	}
}

func mapgenPlaceRuntimeDeps503B30() mapgenPlaceDeps503B30 {
	return mapgenPlaceDeps503B30{
		fixCoords: func(at mapgenPlacePoint503B30) (mapgenPlacePoint503B30, bool) {
			var x, y C.float
			ok := C.nox_mapgenFixCoordsNative_503B30(
				C.float(at.x), C.float(at.y), &x, &y,
			) != 0
			return mapgenPlacePoint503B30{x: float32(x), y: float32(y)}, ok
		},
		placementReady: func() bool {
			loaded := uint32(C.dword_5d4594_1599480)
			return loaded == uint32(C.dword_5d4594_3835396) && int32(loaded) != -1 &&
				uint32(C.dword_5d4594_1599476) != 1
		},
		loadSelected: func() bool {
			return C.nox_xxx_mapgenSaveMap_503830(C.int(int32(C.dword_5d4594_3835396))) != 0
		},
		geometry: mapgenPlaceRuntimeGeometry503B30,
		storeGeometry: func(geometry mapgenPlaceGeometry503B30) {
			*memmap.PtrUint32(0x5D4594, 1599484) = geometry.wallSize[0]
			*memmap.PtrUint32(0x5D4594, 1599488) = geometry.wallSize[1]
			*memmap.PtrFloat32(0x5D4594, 1599492) = geometry.wallSpan[0]
			*memmap.PtrFloat32(0x5D4594, 1599496) = geometry.wallSpan[1]
		},
		placeTiles: func(delta [2]int32) bool {
			return C.nox_xxx_tileInit_504150(C.int(delta[0]), C.int(delta[1])) != 0
		},
		placeWalls: func(delta [2]int32) bool {
			return C.sub_504330(C.int32_t(delta[0]), C.int32_t(delta[1])) != 0
		},
		placeWaypoints: func(delta [2]int32) bool {
			return C.sub_504560(C.int32_t(delta[0]), C.int32_t(delta[1])) != 0
		},
		placeObjects: func(delta [2]int32) bool {
			return C.sub_504910(C.int32_t(delta[0]), C.int32_t(delta[1])) != 0
		},
		prepareWaypoints: func() {
			GetServer().S().WPs.Sub_579D20()
		},
		markPendingWaypoints: func() {
			for wp := GetServer().S().WPs.Pending; wp != nil; wp = wp.WpNext {
				wp.Flags |= 0x80000000
			}
		},
		fixupPendingObjects: func(bounds [4]int32) {
			origin := [2]uint32{uint32(bounds[0]), uint32(bounds[1])}
			next := interestingXferNative4D0010(
				&origin, int32(C.dword_5d4594_3835392), interestingXferRuntimeDeps4D0010(),
			)
			C.dword_5d4594_3835392 = C.uint32_t(uint32(next))
		},
		placeGroups: func(delta [2]int32) bool {
			return GetServer().Sub504720(uint32(delta[0]), uint32(delta[1])) != 0
		},
		clearWaypointTempIDs: func() {
			for wp := GetServer().S().WPs.Pending; wp != nil; wp = wp.WpNext {
				wp.Field1 = 0
			}
		},
		clearObjectScriptIDs: func() {
			for obj := GetServer().S().Objs.Pending; obj != nil; obj = obj.ObjNext {
				obj.ScriptIDVal = 0
			}
		},
		finalizeWaypoints: func() {
			GetServer().S().Nox_xxx_waypoint_5799C0()
		},
		finalizePendingObjects: func() {
			GetServer().S().Objs.ObjectsClearPending()
		},
		markPlaced: func() {
			C.dword_5d4594_1599476 = 1
		},
		hasScriptPayload: func() bool {
			return C.dword_5d4594_1599644 != 0
		},
		finishScriptPayload: mapgenFinishScriptPayload503B30,
		advancePlacementIndex: func() {
			C.dword_5d4594_3835312++
		},
	}
}

var mapgenPlaceEntry503B30 = func(x, y float32) bool {
	return mapgenPlaceWithDeps503B30(
		mapgenPlacePoint503B30{x: x, y: y}, mapgenPlaceRuntimeDeps503B30(),
	)
}

//export nox_mapgenPlaceNative_503B30
func nox_mapgenPlaceNative_503B30(x, y C.float) C.int {
	return C.int(bool2int(mapgenPlaceEntry503B30(float32(x), float32(y))))
}
