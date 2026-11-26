package legacy

/*
#include "GAME4_2.h"
#include "xfer_door_4f4cb0.h"

*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type mapgenObjectPlacementAccess struct {
	fix          func(*mapgenWorldPoint, *mapgenWorldPoint) bool
	allocate     func(int32) *server.Object
	usesDoorXfer func(*server.Object) bool
	create       func(*server.Object)
	clearPending func()
}

func mapgenSelectObjectType(name string, selected *uint32, resolve func() int32) {
	if strings.EqualFold(name, "NONE") {
		*selected = 0
		return
	}
	*selected = uint32(resolve())
}

func mapgenPlaceObject(position *mapgenWorldPoint, selected uint32, extent *uint32, access mapgenObjectPlacementAccess) *server.Object {
	if selected == 0 || !access.fix(position, &mapgenWorldPoint{}) {
		return nil
	}
	obj := access.allocate(int32(selected))
	if obj == nil {
		return nil
	}
	return mapgenMoveObject(obj, position, extent, access)
}

func mapgenMoveObject(obj *server.Object, position *mapgenWorldPoint, extent *uint32, access mapgenObjectPlacementAccess) *server.Object {
	var fixed mapgenWorldPoint
	if !access.fix(position, &fixed) || obj == nil {
		return obj
	}

	obj.Extent = *extent
	*extent++
	obj.PosVec = types.Pointf{X: fixed.x, Y: fixed.y}
	obj.ObjFlags |= object.FlagEnabled
	if access.usesDoorXfer(obj) {
		obj.PosVec.X = mapgenSnapDoorCoordinate(fixed.x)
		obj.PosVec.Y = mapgenSnapDoorCoordinate(fixed.y)
	}
	access.create(obj)
	access.clearPending()
	return obj
}

func mapgenSnapDoorCoordinate(value float32) float32 {
	scaled := mulFloat64TowardZero(float64(value), mapgenWallCoordinateScale)
	cell := x87TruncQwordLow(addFloat64TowardZero(scaled, 0.5))
	return storeFloat32TowardZero(float64(cell * 23))
}

func mapgenObjectPlacementProductionAccess() mapgenObjectPlacementAccess {
	return mapgenObjectPlacementAccess{
		fix: mapgenFixWorldCoords,
		allocate: func(index int32) *server.Object {
			return asObjectS(nox_xxx_newObjectWithTypeInd_4E3450(index))
		},
		usesDoorXfer: func(obj *server.Object) bool {
			return nox_objectTypeGetXfer(nox_xxx_getUnitName_4E39D0(asObjectC(obj))) == C.nox_xxx_XFerDoor_4F4CB0
		},
		create: func(obj *server.Object) {
			nox_xxx_createAt_4DAA50(asObjectC(obj), nil, C.float(obj.PosVec.X), C.float(obj.PosVec.Y))
		},
		clearPending: nox_xxx_unitClearPendingMB_4DB030,
	}
}

//export nox_xxx_mapGenGetObjID_527940
func nox_xxx_mapGenGetObjID_527940(name *C.char) C.int {
	if name == nil {
		return 0
	}
	selected := (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_3835388))
	mapgenSelectObjectType(GoString(name), selected, func() int32 { return nox_xxx_getNameId_4E3AA0(name) })
	return 1
}

//export nox_xxx_mapGenPlaceObj_5279B0
func nox_xxx_mapGenPlaceObj_5279B0(position *C.float2) *C.nox_object_t {
	extent := (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_3835392))
	return asObjectC(mapgenPlaceObject((*mapgenWorldPoint)(unsafe.Pointer(position)),
		uint32(legacyAddressableWords.dword_5d4594_3835388), extent, mapgenObjectPlacementProductionAccess()))
}

//export nox_xxx_mapGenMoveObject_527A10
func nox_xxx_mapGenMoveObject_527A10(obj *C.nox_object_t, position *C.float2) *C.nox_object_t {
	extent := (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_3835392))
	return asObjectC(mapgenMoveObject(asObjectS(obj), (*mapgenWorldPoint)(unsafe.Pointer(position)),
		extent, mapgenObjectPlacementProductionAccess()))
}
