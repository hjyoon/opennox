package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

func mapgenHallMidpoint(first, last C.uint32_t) C.float {
	sum := float32(math.Float32frombits(uint32(last)) + math.Float32frombits(uint32(first)))
	return C.float(float64(sum) * 0.5)
}

func mapgenHallPointOnEdge(room *C.uint8_t, point *C.float2, far bool) C.int {
	result := C.int(registerMapgenPointer(unsafe.Pointer(room)))
	if room == nil || point == nil {
		return result
	}
	hall := (*mapgenRoomGeometry)(unsafe.Pointer(room))
	switch hall.kind {
	case 2, 3:
		edge := 1
		if far && hall.kind == 2 || !far && hall.kind == 3 {
			edge = 3
		}
		point.field_0 = mapgenHallMidpoint(hall.bounds[0], hall.bounds[2])
		point.field_4 = C.float(math.Float32frombits(uint32(hall.bounds[edge])))
		if hall.kind == 2 {
			result = C.int(hall.bounds[edge])
		}
	case 4, 5:
		edge := 0
		if far && hall.kind == 5 || !far && hall.kind == 4 {
			edge = 2
		}
		point.field_0 = C.float(math.Float32frombits(uint32(hall.bounds[edge])))
		point.field_4 = mapgenHallMidpoint(hall.bounds[1], hall.bounds[3])
		result = C.int(hall.bounds[edge])
	}
	return result
}

//export nox_mapgenConnectionPointNearNative_523C30
func nox_mapgenConnectionPointNearNative_523C30(room *C.uint8_t, point *C.float2) C.int {
	return mapgenHallPointOnEdge(room, point, false)
}

//export sub_523C30
func sub_523C30(room, point C.int) C.int {
	raw := mapgenResolveListedRoom(uint32(room))
	output := resolveMapgenPointer(uintptr(uint32(point)))
	return nox_mapgenConnectionPointNearNative_523C30((*C.uint8_t)(unsafe.Pointer(raw)), (*C.float2)(output))
}

//export nox_mapgenConnectionPointFarNative_523CB0
func nox_mapgenConnectionPointFarNative_523CB0(room *C.uint8_t, point *C.float2) C.int {
	return mapgenHallPointOnEdge(room, point, true)
}

//export sub_523CB0
func sub_523CB0(room, point C.int) C.int {
	raw := mapgenResolveListedRoom(uint32(room))
	output := resolveMapgenPointer(uintptr(uint32(point)))
	return nox_mapgenConnectionPointFarNative_523CB0((*C.uint8_t)(unsafe.Pointer(raw)), (*C.float2)(output))
}

func mapgenHallPointCenter(hall *mapgenRoomGeometry, point *C.float2) {
	links := (*mapgenRoomLinkRecord)(unsafe.Pointer(hall))
	direction := 2
	if hall.kind == 2 || hall.kind == 3 {
		direction = 0
	}
	first := mapgenResolveRoomLinkRecord(links.links[direction][0])
	second := mapgenResolveRoomLinkRecord(links.links[direction+1][0])
	if first != nil && second != nil && first.kind == 1 && second.kind == 1 {
		point.field_0 = mapgenHallMidpoint(hall.bounds[0], hall.bounds[2])
		point.field_4 = mapgenHallMidpoint(hall.bounds[1], hall.bounds[3])
		return
	}

	bound := func(index int) float64 { return float64(math.Float32frombits(uint32(hall.bounds[index]))) }
	width := func(index int) float64 { return float64(math.Float32frombits(uint32(hall.dimensions[index]))) }
	switch hall.kind {
	case 2:
		point.field_0 = mapgenHallMidpoint(hall.bounds[0], hall.bounds[2])
		point.field_4 = C.float(bound(1) + width(0)*0.5)
	case 3:
		point.field_0 = mapgenHallMidpoint(hall.bounds[0], hall.bounds[2])
		point.field_4 = C.float(bound(3) - width(0)*0.5)
	case 4:
		point.field_0 = C.float(bound(2) - width(1)*0.5)
		point.field_4 = mapgenHallMidpoint(hall.bounds[1], hall.bounds[3])
	case 5:
		point.field_0 = C.float(bound(0) + width(1)*0.5)
		point.field_4 = mapgenHallMidpoint(hall.bounds[1], hall.bounds[3])
	}
}

//export nox_mapgenConnectionPointCenterNative_523D30
func nox_mapgenConnectionPointCenterNative_523D30(room *C.uint8_t, point *C.float2) *C.float {
	if room != nil && point != nil {
		mapgenHallPointCenter((*mapgenRoomGeometry)(unsafe.Pointer(room)), point)
	}
	return (*C.float)(unsafe.Pointer(room))
}

//export sub_523D30
func sub_523D30(room, point *C.float) *C.float {
	raw := resolveMapgenPointer(uintptr(unsafe.Pointer(room)))
	output := resolveMapgenPointer(uintptr(unsafe.Pointer(point)))
	return nox_mapgenConnectionPointCenterNative_523D30((*C.uint8_t)(raw), (*C.float2)(output))
}

func mapgenShortenHall(hall, target *mapgenRoomGeometry) bool {
	axis, across := 0, 1
	if hall.kind == 2 || hall.kind == 3 {
		axis, across = 1, 0
	}
	asFloat := func(bits C.uint32_t) float64 { return float64(math.Float32frombits(uint32(bits))) }
	if asFloat(hall.bounds[across]) < asFloat(target.bounds[across]) ||
		asFloat(hall.bounds[across+2]) > asFloat(target.bounds[across+2]) {
		return false
	}
	pos := asFloat(hall.position[axis])
	minimum, maximum := asFloat(target.bounds[axis]), asFloat(target.bounds[axis+2])
	for hall.cellDimensions[axis] > 1 {
		size := asFloat(hall.dimensions[axis]) - mapgenHallCellSize
		hall.cellDimensions[axis]--
		hall.dimensions[axis] = C.uint32_t(math.Float32bits(float32(size)))
		var keepShrinking bool
		switch hall.kind {
		case 2, 5:
			pos += mapgenHallCellSize
			keepShrinking = pos+0.1 <= maximum
		case 3:
			keepShrinking = !(size+pos-0.1 < minimum)
		case 4:
			keepShrinking = size+pos-0.1 >= minimum
		}
		if !keepShrinking {
			break
		}
	}
	newPos := hall.position
	newPos[axis] = C.uint32_t(math.Float32bits(float32(pos)))
	nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(unsafe.Pointer(hall)), (*C.float2)(unsafe.Pointer(&newPos[0])))
	return true
}

//export nox_mapgenAdjustHallNative_523A10
func nox_mapgenAdjustHallNative_523A10(room *C.uint8_t, target *C.nox_mapgen_readonly_float_t) C.int {
	if room == nil || target == nil {
		return 0
	}
	hall := (*mapgenRoomGeometry)(unsafe.Pointer(room))
	if mapgenIsHall(uint32(hall.kind)) && !mapgenShortenHall(hall, (*mapgenRoomGeometry)(unsafe.Pointer(target))) {
		return 0
	}
	return C.int(bool2int(nox_mapgenRoomAtNative_521200(room) == nil))
}

//export sub_523A10
func sub_523A10(room C.int, target *C.float) C.int {
	raw := mapgenResolveListedRoom(uint32(room))
	return nox_mapgenAdjustHallNative_523A10((*C.uint8_t)(unsafe.Pointer(raw)), (*C.nox_mapgen_readonly_float_t)(unsafe.Pointer(target)))
}
