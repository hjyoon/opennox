package legacy

/*
#include "GAME4_2.h"
#include "server__mapgen__generate__populate.h"
#include "xfer_exit_4f4b90.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type mapgenExitRoom struct {
	mapgenRoomGeometry
	_       [216 - unsafe.Sizeof(mapgenRoomGeometry{})]byte
	blocked [4]byte
}

type mapgenExitDefinition struct {
	name [60]byte
	side uint32
}

type mapgenExitTheme struct {
	_           [216]byte
	exits       [4]mapgenExitDefinition
	exitCount   int32
	destination [60]byte
}

var (
	_ [216]byte = [unsafe.Offsetof(mapgenExitRoom{}.blocked)]byte{}
	_ [64]byte  = [unsafe.Sizeof(mapgenExitDefinition{})]byte{}
	_ [472]byte = [unsafe.Offsetof(mapgenExitTheme{}.exitCount)]byte{}
	_ [476]byte = [unsafe.Offsetof(mapgenExitTheme{}.destination)]byte{}
)

type mapgenExitPlacementAccess struct {
	selectType func(*byte)
	spawn      func(*mapgenWorldPoint) *server.Object
	round      func(*mapgenWorldPoint) (int32, int32)
	occupy     func(*mapgenExitRoom, *mapgenWorldPoint, float32, float32)
}

func mapgenExitCenter(dimension, edge float64) float64 {
	return addFloat64TowardZero(addFloat64TowardZero(dimension*0.5, edge), 1)
}

func mapgenPlaceExit(room *mapgenExitRoom, definition *mapgenExitDefinition, access mapgenExitPlacementAccess) *server.Object {
	if room == nil || definition == nil {
		return nil
	}
	side := definition.side
	if side >= uint32(len(room.blocked)) || room.blocked[side] != 0 {
		return nil
	}
	left := float64(math.Float32frombits(uint32(room.bounds[0])))
	top := float64(math.Float32frombits(uint32(room.bounds[1])))
	right := float64(math.Float32frombits(uint32(room.bounds[2])))
	bottom := float64(math.Float32frombits(uint32(room.bounds[3])))
	width := float64(math.Float32frombits(uint32(room.dimensions[0])))
	height := float64(math.Float32frombits(uint32(room.dimensions[1])))
	var x, y float64
	switch side {
	case 0:
		x, y = mapgenExitCenter(width, left), addFloat64TowardZero(top, 10)
	case 1:
		x, y = mapgenExitCenter(width, left), addFloat64TowardZero(bottom, -10)
	case 2:
		x, y = addFloat64TowardZero(right, -10), mapgenExitCenter(height, top)
	case 3:
		x, y = addFloat64TowardZero(left, 10), mapgenExitCenter(height, top)
	}
	position := mapgenWorldPoint{x: storeFloat32TowardZero(x), y: storeFloat32TowardZero(y)}
	access.selectType(&definition.name[0])
	object := access.spawn(&position)
	if object == nil {
		return nil
	}

	col, row := access.round(&position)
	var columns, rows int32
	switch definition.side {
	case 0:
		col--
		columns, rows = 3, 2
	case 1:
		col--
		row -= 2
		columns, rows = 3, 2
	case 2:
		col -= 2
		row--
		columns, rows = 2, 3
	case 3:
		row--
		columns, rows = 2, 3
	}
	origin := mapgenWorldPoint{x: mapgenExitCellToWorld(col), y: mapgenExitCellToWorld(row)}
	access.occupy(room, &origin, mapgenExitCellToWorld(columns), mapgenExitCellToWorld(rows))
	return object
}

func mapgenExitCellToWorld(cells int32) float32 {
	return storeFloat32TowardZero(mulFloat64TowardZero(float64(cells), mapgenCellSize))
}

func mapgenSetExitDestination(object *server.Object, destination *byte, usesExitXfer func(*server.Object) bool) bool {
	if object == nil || destination == nil || object.CollideData == nil || !usesExitXfer(object) {
		return false
	}
	name := &(*server.ExitCollideData)(object.CollideData).MapName
	for index := range name {
		value := *(*byte)(unsafe.Add(unsafe.Pointer(destination), index))
		if value == 0 {
			clear(name[index:])
			break
		}
		name[index] = value
	}
	return true
}

type mapgenMakeExitAccess struct {
	farthest       func() *mapgenExitRoom
	previous       func(*mapgenExitRoom) *mapgenExitRoom
	place          func(*mapgenExitRoom, *mapgenExitDefinition) *server.Object
	setDestination func(*server.Object, *byte) bool
}

func mapgenMakeExit(theme *mapgenExitTheme, access mapgenMakeExitAccess) bool {
	if theme == nil || theme.exitCount == 0 {
		return true
	}
	for room := access.farthest(); room != nil; room = access.previous(room) {
		if room.kind != 1 {
			continue
		}
		for index := int32(0); index < theme.exitCount && index < int32(len(theme.exits)); index++ {
			object := access.place(room, &theme.exits[index])
			if object != nil {
				access.setDestination(object, &theme.destination[0])
				return true
			}
		}
	}
	return false
}

func mapgenPlaceExitProduction(room *mapgenExitRoom, definition *mapgenExitDefinition) *server.Object {
	return mapgenPlaceExit(room, definition, mapgenExitPlacementAccess{
		selectType: func(name *byte) {
			nox_xxx_mapGenGetObjID_527940((*C.char)(unsafe.Pointer(name)))
		},
		spawn: func(position *mapgenWorldPoint) *server.Object {
			return asObjectS(nox_xxx_mapGenPlaceObj_5279B0((*C.float2)(unsafe.Pointer(position))))
		},
		round: func(position *mapgenWorldPoint) (int32, int32) {
			return int32(mapgenGridCoordinate(position.x)), int32(mapgenGridCoordinate(position.y))
		},
		occupy: func(room *mapgenExitRoom, position *mapgenWorldPoint, width, height float32) {
			nox_mapgenAddOccupiedRectNative_521BC0((*C.uint8_t)(unsafe.Pointer(room)),
				(*C.float2)(unsafe.Pointer(position)), C.float(width), C.float(height))
		},
	})
}

func mapgenSetExitDestinationProduction(object *server.Object, destination *byte) bool {
	return mapgenSetExitDestination(object, destination, func(object *server.Object) bool {
		name := nox_xxx_getUnitName_4E39D0(asObjectC(object))
		return nox_objectTypeGetXfer(name) == C.nox_xxx_XFerExit_4F4B90
	})
}

//export nox_mapgenMakeExitNative_522A40
func nox_mapgenMakeExitNative_522A40(theme *C.uint8_t) C.int {
	return C.int(bool2int(mapgenMakeExit((*mapgenExitTheme)(unsafe.Pointer(theme)), mapgenMakeExitAccess{
		farthest: func() *mapgenExitRoom {
			return (*mapgenExitRoom)(unsafe.Pointer(nox_mapgenFarthestRoomNative_5259E0()))
		},
		previous: func(room *mapgenExitRoom) *mapgenExitRoom {
			return (*mapgenExitRoom)(unsafe.Pointer(nox_mapgenRoomSortedPrevNative_525C90((*C.nox_mapgen_readonly_room_t)(unsafe.Pointer(room)))))
		},
		place:          mapgenPlaceExitProduction,
		setDestination: mapgenSetExitDestinationProduction,
	})))
}

//export nox_xxx_mapGenMakeExit_522A40
func nox_xxx_mapGenMakeExit_522A40(themeToken C.int) C.int {
	theme := resolveMapgenPointer(uintptr(uint32(themeToken)))
	return nox_mapgenMakeExitNative_522A40((*C.uint8_t)(theme))
}

//export nox_xxx_mapgen_522AD0
func nox_xxx_mapgen_522AD0(roomRef *C.float, definitionToken C.int) *C.float {
	room := resolveMapgenPointer(uintptr(unsafe.Pointer(roomRef)))
	definition := resolveMapgenPointer(uintptr(uint32(definitionToken)))
	object := mapgenPlaceExitProduction((*mapgenExitRoom)(room), (*mapgenExitDefinition)(definition))
	return (*C.float)(unsafe.Pointer(object))
}

//export sub_527D50
func sub_527D50(objectToken C.int, destination *C.char) C.int {
	object := resolveMapgenPointer(uintptr(uint32(objectToken)))
	return C.int(bool2int(mapgenSetExitDestinationProduction((*server.Object)(object), (*byte)(unsafe.Pointer(destination)))))
}
