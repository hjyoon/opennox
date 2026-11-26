package legacy

/*
#include <stdlib.h>
#include "GAME1_1.h"

*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	mapPolygonTableBase = 552228
	mapAngleTableBase   = 535844
)

type mapAngleRemap struct {
	index, original uint32
	next            *mapAngleRemap
}

var (
	mapPolygonScriptData [mapPolygonCapacity]unsafe.Pointer
	mapPolygonVertexData [mapPolygonCapacity]*C.uint32_t
)

func mapPolygonIndex(polygon unsafe.Pointer) (int, bool) {
	start := uintptr(memmap.PtrOff(0x5D4594, mapPolygonTableBase))
	offset := uintptr(polygon) - start
	size := unsafe.Sizeof(mapPolygonRecord{})
	index := offset / size
	return int(index), offset%size == 0 && index < mapPolygonCapacity
}

func mapPolygonByIndex(index uint32) *mapPolygonRecord {
	return memmap.PtrT[mapPolygonRecord](0x5D4594, uintptr(mapPolygonTableBase+int32(index)*int32(unsafe.Sizeof(mapPolygonRecord{}))))
}

func mapAngleByIndex(index uint32) *Nox_polygon_angle_data {
	return memmap.PtrT[Nox_polygon_angle_data](0x5D4594, uintptr(mapAngleTableBase+int32(index)*int32(unsafe.Sizeof(Nox_polygon_angle_data{}))))
}

//export nox_xxx_polygonGetDataNative
func nox_xxx_polygonGetDataNative(polygon unsafe.Pointer) unsafe.Pointer {
	if index, ok := mapPolygonIndex(polygon); ok {
		return mapPolygonScriptData[index]
	}
	return nil
}

//export nox_xxx_polygonSetDataNative
func nox_xxx_polygonSetDataNative(polygon, data unsafe.Pointer) {
	if index, ok := mapPolygonIndex(polygon); ok {
		mapPolygonScriptData[index] = data
		(*mapPolygonRecord)(polygon).DataMarker = 0
	}
}

//export nox_xxx_polygonGetVertexIndicesNative
func nox_xxx_polygonGetVertexIndicesNative(polygon unsafe.Pointer) *C.uint32_t {
	if index, ok := mapPolygonIndex(polygon); ok {
		return mapPolygonVertexData[index]
	}
	return nil
}

//export nox_xxx_polygonSetVertexIndicesNative
func nox_xxx_polygonSetVertexIndicesNative(polygon unsafe.Pointer, vertices *C.uint32_t) {
	if index, ok := mapPolygonIndex(polygon); ok {
		mapPolygonVertexData[index] = vertices
		(*mapPolygonRecord)(polygon).VertexMarker = 0
	}
}

//export sub_420C40
func sub_420C40(index, original C.int) *C.uint32_t {
	p := (*mapAngleRemap)(C.calloc(1, C.size_t(unsafe.Sizeof(mapAngleRemap{}))))
	if p == nil {
		return nil
	}
	p.index, p.original = uint32(index), uint32(original)
	p.next = (*mapAngleRemap)(legacyPointers.polygonAngleRemapHead)
	legacyPointers.polygonAngleRemapHead = unsafe.Pointer(p)
	return (*C.uint32_t)(unsafe.Pointer(p))
}

//export sub_420C70
func sub_420C70() *C.uint32_t {
	head := (*mapAngleRemap)(legacyPointers.polygonAngleRemapHead)
	for p := head; p != nil; {
		next := p.next
		C.free(unsafe.Pointer(p))
		p = next
	}
	legacyPointers.polygonAngleRemapHead = nil
	return (*C.uint32_t)(unsafe.Pointer(head))
}

func mapAngleNext(index uint32) *Nox_polygon_angle_data {
	for limit := uint32(legacyWords.nox_xxx_polygonNextAngle_587000_60356); index < limit; index++ {
		angle := mapAngleByIndex(index)
		if angle.Active != 0 {
			return angle
		}
	}
	return nil
}

//export nox_xxx_polygon_420CA0
func nox_xxx_polygon_420CA0() *C.char {
	return (*C.char)(unsafe.Pointer(mapAngleNext(1)))
}

//export nox_xxx_polygon_420CD0
func nox_xxx_polygon_420CD0(index *C.uint32_t) *C.char {
	return (*C.char)(unsafe.Pointer(mapAngleNext(uint32(*index) + 1)))
}

//export sub_420D10
func sub_420D10() C.int {
	limit := legacyWords.nox_xxx_polygonNextAngle_587000_60356
	for index := uint32(1); index < uint32(limit); index++ {
		if mapAngleByIndex(index).Active == 0 {
			return C.int(index)
		}
	}
	legacyWords.nox_xxx_polygonNextAngle_587000_60356++
	return C.int(limit)
}

//export nox_xxx_polygonSetAngle_420D40
func nox_xxx_polygonSetAngle_420D40(x, y C.int, index C.uint, remap C.int) *C.uint {
	if remap != 0 {
		sub_420C40(C.int(index), remap)
	}
	angle := mapAngleByIndex(uint32(index))
	angle.Index, angle.Active = uint32(index), 1
	angle.X, angle.Y = math.Float32frombits(uint32(x)), math.Float32frombits(uint32(y))
	if C.uint32_t(index) >= legacyWords.nox_xxx_polygonNextAngle_587000_60356 {
		legacyWords.nox_xxx_polygonNextAngle_587000_60356 = C.uint32_t(index) + 1
	}
	return (*C.uint)(unsafe.Pointer(angle))
}

//export sub_421010
func sub_421010() *C.char {
	for index := uint32(1); index < mapPolygonAngleCapacity; index++ {
		mapAngleByIndex(index).Active = 0
	}
	legacyWords.nox_xxx_polygonNextAngle_587000_60356 = 1
	firstPolygon := mapPolygonByIndex(0)
	return (*C.char)(unsafe.Add(unsafe.Pointer(firstPolygon), 12))
}

//export nox_xxx_polygonGetAngle_421030
func nox_xxx_polygonGetAngle_421030(index C.int) *C.char {
	return (*C.char)(unsafe.Pointer(mapAngleByIndex(uint32(index))))
}

//export sub_421040
func sub_421040(polygon unsafe.Pointer) {
	data := nox_xxx_polygonGetVertexIndicesNative(polygon)
	if data == nil {
		return
	}
	vertices := unsafe.Slice(data, (*mapPolygonRecord)(polygon).VertexCount)
	for i, index := range vertices {
		for p := (*mapAngleRemap)(legacyPointers.polygonAngleRemapHead); p != nil; p = p.next {
			if uint32(index) == p.original {
				vertices[i] = C.uint32_t(p.index)
				break
			}
		}
	}
}

func mapPolygonNext(index uint32) *mapPolygonRecord {
	for limit := uint32(worldWordStorage.next_polygon_index); index < limit; index++ {
		polygon := mapPolygonByIndex(index)
		if polygon.Active != 0 {
			return polygon
		}
	}
	return nil
}

//export nox_xxx_polygonGetNext_4210A0
func nox_xxx_polygonGetNext_4210A0() *C.char {
	return (*C.char)(unsafe.Pointer(mapPolygonNext(1)))
}

//export sub_4210E0
func sub_4210E0(previous unsafe.Pointer) *C.char {
	return (*C.char)(unsafe.Pointer(mapPolygonNext((*mapPolygonRecord)(previous).Index + 1)))
}

//export sub_421130
func sub_421130() C.int {
	limit := worldWordStorage.next_polygon_index
	for index := uint32(1); index < uint32(limit); index++ {
		if mapPolygonByIndex(index).Active == 0 {
			return C.int(index)
		}
	}
	worldWordStorage.next_polygon_index++
	return C.int(limit)
}

//export nox_xxx_polygonGetByIdx_4214A0
func nox_xxx_polygonGetByIdx_4214A0(index C.int) *C.char {
	if uint32(index) == 0xdeadface {
		return nil
	}
	return (*C.char)(unsafe.Pointer(mapPolygonByIndex(uint32(index))))
}

//export sub_421430
func sub_421430() unsafe.Pointer {
	var result unsafe.Pointer
	release := memmap.Uint32(0x5D4594, 588076) != 0
	for index := range mapPolygonCapacity {
		polygon := mapPolygonByIndex(uint32(index))
		address := unsafe.Pointer(polygon)
		data, vertices := nox_xxx_polygonGetDataNative(address), nox_xxx_polygonGetVertexIndicesNative(address)
		result = unsafe.Pointer(vertices)
		if release {
			C.free(data)
			C.free(result)
		}
		nox_xxx_polygonSetDataNative(address, nil)
		nox_xxx_polygonSetVertexIndicesNative(address, nil)
		polygon.VertexCount, polygon.Index, polygon.Active = 0, 0, 0
		polygon.Handlers[0].Func, polygon.Handlers[1].Func = -1, -1
	}
	worldWordStorage.next_polygon_index = 1
	return result
}
