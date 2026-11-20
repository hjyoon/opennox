package legacy

/*
#include "GAME4_2.h"
#include "GAME3_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

const mapGenMonsterSpread522810 float32 = 0.95

func mapGenPointOccupied5227B0(head uint32, point *types.Pointf, resolve func(uint32) *mapGenOccupiedRect5227B0) bool {
	for head != 0 {
		rect := resolve(head)
		if rect == nil {
			return false
		}
		b := rect.bounds
		// The original upper-bound branch accepts an unordered comparison.
		if point.X >= b[0] && !(point.X > b[2]) && point.Y >= b[1] && !(point.Y > b[3]) {
			return true
		}
		head = rect.next
	}
	return false
}

func mapGenPointOffset5226D0(offset float64, scale, center float32) float32 {
	scaled := mulFloat64TowardZero(offset, float64(scale))
	return storeFloat32TowardZero(addFloat64TowardZero(scaled, float64(center)))
}

func mapGenRoomHalfSum5226D0(a, b float32) float32 {
	value := addFloat64TowardZero(float64(a), float64(b))
	return storeFloat32TowardZero(value * 0.5)
}

func mapGenRoomHalfDifference5226D0(a, b float32) float32 {
	// x87 subtraction propagates the NaN operand with its sign unchanged.
	if !math.IsNaN(float64(b)) {
		b = -b
	}
	return mapGenRoomHalfSum5226D0(a, b)
}

func mapGenFindFreePoint5226D0(bounds [4]float32, scale float32, point *types.Pointf,
	random func(float32, float32) float64, occupied func(*types.Pointf) bool) bool {
	centerX := mapGenRoomHalfSum5226D0(bounds[0], bounds[2])
	centerY := mapGenRoomHalfSum5226D0(bounds[1], bounds[3])
	halfWidth := mapGenRoomHalfDifference5226D0(bounds[2], bounds[0])
	halfHeight := mapGenRoomHalfDifference5226D0(bounds[3], bounds[1])
	for i := 0; i < 10; i++ {
		point.X = mapGenPointOffset5226D0(random(-halfWidth, halfWidth), scale, centerX)
		point.Y = mapGenPointOffset5226D0(random(-halfHeight, halfHeight), scale, centerY)
		if !occupied(point) {
			return true
		}
	}
	return false
}

type mapGenRoomMonsterRuntime522810 struct {
	find           func(float32, *types.Pointf) bool
	selectType     func()
	place          func(*types.Pointf) *server.Object
	angle          func(*types.Pointf) int
	directionIndex func(int) int
	orient         func(*server.Object, int)
}

func mapGenSpawnRoomMonster522810(bounds [4]float32, rt mapGenRoomMonsterRuntime522810) *server.Object {
	centerX := mapGenRoomHalfSum5226D0(bounds[2], bounds[0])
	centerY := mapGenRoomHalfSum5226D0(bounds[3], bounds[1])
	var point types.Pointf
	if !rt.find(mapGenMonsterSpread522810, &point) {
		return nil
	}
	rt.selectType()
	spawned := rt.place(&point)
	if spawned != nil && spawned.Class().Has(object.ClassMonster) {
		direction := types.Pointf{
			X: storeFloat32TowardZero(addFloat64TowardZero(float64(centerX), -float64(point.X))),
			Y: storeFloat32TowardZero(addFloat64TowardZero(float64(centerY), -float64(point.Y))),
		}
		rt.orient(spawned, rt.directionIndex(rt.angle(&direction)))
	}
	return spawned
}

func mapGenOrientObject527C60(unit *server.Object, direction int32,
	reduce func(int32) int32, angle func(int32) int32) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	if unit.Class().Has(object.ClassMonster) {
		update := unit.UpdateDataMonster()
		facing := angle(reduce(direction))
		update.Direction94 = uint32(facing)
		unit.Direction1 = server.Dir16(facing)
		return true
	}
	if !unit.Class().Has(object.ClassDoor) {
		return false
	}
	var facing int32
	switch direction {
	case 1:
		facing = 0
	case 3:
		facing = 24
	case 5:
		facing = 8
	case 7:
		facing = 16
	default:
		return false
	}
	update := unit.UpdateDataDoor()
	update.CurrentDirection = facing
	update.TargetDirection = facing
	update.SyncedDirection = facing
	return true
}

func mapGenRoomBounds5226D0(room *C.uint8_t) [4]float32 {
	raw := (*mapgenRoomGeometry)(unsafe.Pointer(room)).bounds
	var bounds [4]float32
	for i, value := range raw {
		bounds[i] = math.Float32frombits(uint32(value))
	}
	return bounds
}

func mapGenRoomOccupied5227B0(room *C.uint8_t, point *types.Pointf) bool {
	return mapGenPointOccupied5227B0(*mapGenOccupiedRectHead(room), point, mapGenResolveOccupiedRect)
}

//export nox_mapgenPointOccupiedNative_5227B0
func nox_mapgenPointOccupiedNative_5227B0(room *C.uint8_t, point *C.float) C.int {
	if room == nil || point == nil {
		return 0
	}
	return C.int(bool2int(mapGenRoomOccupied5227B0(room, (*types.Pointf)(unsafe.Pointer(point)))))
}

//export sub_5227B0
func sub_5227B0(room *C.uint8_t, point *C.float) C.int {
	return nox_mapgenPointOccupiedNative_5227B0(room, point)
}

//export sub_5226D0
func sub_5226D0(room *C.uint8_t, scale C.float, point *C.float2) C.int {
	if room == nil || point == nil {
		return 0
	}
	return C.int(bool2int(mapGenFindFreePoint5226D0(
		mapGenRoomBounds5226D0(room), float32(scale), (*types.Pointf)(unsafe.Pointer(point)),
		func(lo, hi float32) float64 { return float64(C.sub_526BC0(C.float(lo), C.float(hi))) },
		func(point *types.Pointf) bool { return mapGenRoomOccupied5227B0(room, point) },
	)))
}

//export nox_xxx_mapGenMakeMonsterInRoom_522810
func nox_xxx_mapGenMakeMonsterInRoom_522810(room *C.uint8_t, name *C.char) *C.nox_object_t {
	if room == nil || name == nil {
		return nil
	}
	spawned := mapGenSpawnRoomMonster522810(mapGenRoomBounds5226D0(room), mapGenRoomMonsterRuntime522810{
		find: func(scale float32, point *types.Pointf) bool {
			return sub_5226D0(room, C.float(scale), (*C.float2)(unsafe.Pointer(point))) != 0
		},
		selectType: func() { nox_xxx_mapGenGetObjID_527940(name) },
		place: func(point *types.Pointf) *server.Object {
			return asObjectS(nox_xxx_mapGenPlaceObj_5279B0((*C.float2)(unsafe.Pointer(point))))
		},
		angle: func(direction *types.Pointf) int {
			return int(nox_xxx_math_509ED0((*C.float2)(unsafe.Pointer(direction))))
		},
		directionIndex: func(angle int) int { return int(nox_xxx_math_509EA0(C.int(angle))) },
		orient: func(unit *server.Object, direction int) {
			nox_mapgenOrientObjNative_527C60(asObjectC(unit), C.int(direction))
		},
	})
	return asObjectC(spawned)
}

//export nox_mapgenOrientObjNative_527C60
func nox_mapgenOrientObjNative_527C60(unit *C.nox_object_t, direction C.int) C.int {
	return C.int(bool2int(mapGenOrientObject527C60(asObjectS(unit), int32(direction),
		func(direction int32) int32 { return int32(C.sub_4D3FF0(C.int(direction))) },
		func(index int32) int32 { return int32(nox_xxx_mathDirection4ToAngle_509E90(C.int(index))) },
	)))
}

//export nox_xxx_mapGenOrientObj_527C60
func nox_xxx_mapGenOrientObj_527C60(token, direction C.int) C.int {
	unit := C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(token)))
	return nox_mapgenOrientObjNative_527C60((*C.nox_object_t)(unit), direction)
}
