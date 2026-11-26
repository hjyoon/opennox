package legacy

/*
#include "GAME1_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func mapPolygonVisitEdges(polygon *mapPolygonRecord, visit func(*C.int4) bool) bool {
	data := nox_xxx_polygonGetVertexIndicesNative(unsafe.Pointer(polygon))
	if data == nil || polygon.VertexCount == 0 {
		return false
	}
	vertices := unsafe.Slice(data, polygon.VertexCount)
	coords := func(i int) (C.int, C.int) {
		a := mapAngleByIndex(uint32(vertices[i]))
		return C.int(server.X87TruncateSignedDWORD419A70(a.X)),
			C.int(server.X87TruncateSignedDWORD419A70(a.Y))
	}
	var edge C.int4
	edge.field_0, edge.field_4 = coords(0)
	for i := 1; i <= len(vertices); i++ {
		x, y := coords(i % len(vertices))
		if i&1 != 0 {
			edge.field_8, edge.field_C = x, y
		} else {
			edge.field_0, edge.field_4 = x, y
		}
		if visit(&edge) {
			return true
		}
	}
	return false
}

func mapPolygonContainsPoint(point *[2]int32, polygon *mapPolygonRecord) bool {
	if polygon == nil {
		return false
	}
	frame := memmap.PtrUint32(0x5D4594, 588080)
	ray := C.int4{
		field_0: C.int(point[0]),
		field_4: C.int(point[1]),
		field_8: C.int((*frame & 2) * 0xb80),
	}
	*frame++
	ray.field_C = C.int((*frame & 2) * 0xb80)
	inside := false
	mapPolygonVisitEdges(polygon, func(edge *C.int4) bool {
		if sub_427C80(&ray, edge) != 0 {
			inside = !inside
		}
		return false
	})
	return inside
}

func mapPolygonNearPoint(point *C.int2, polygon *mapPolygonRecord, radius C.float) bool {
	return mapPolygonVisitEdges(polygon, func(edge *C.int4) bool {
		return sub_427DF0(point, &edge.field_0, radius) != 0
	})
}

func findMapPolygon(previous uint32, accepts func(*mapPolygonRecord) bool) *mapPolygonRecord {
	if previous != 0 && previous != 0xdeadface {
		polygon := mapPolygonByIndex(previous)
		if polygon.Active != 0 && accepts(polygon) {
			return polygon
		}
	}
	for index := uint32(1); index < uint32(worldWordStorage.next_polygon_index); index++ {
		polygon := mapPolygonByIndex(index)
		if polygon.Active != 0 && polygon.Index != previous && accepts(polygon) {
			return polygon
		}
	}
	return nil
}

func findMapPolygonAt(point *[2]int32, previous uint32, withHandlers bool) *mapPolygonRecord {
	return findMapPolygon(previous, func(polygon *mapPolygonRecord) bool {
		if withHandlers && polygon.Handlers[0].Func == -1 && polygon.Handlers[1].Func == -1 {
			return false
		}
		if point[0] < polygon.Bounds[0] || point[0] > polygon.Bounds[2] ||
			point[1] < polygon.Bounds[1] || point[1] > polygon.Bounds[3] {
			return false
		}
		return mapPolygonContainsPoint(point, polygon)
	})
}

//export nox_xxx_polygon_421660
func nox_xxx_polygon_421660(point *C.int, polygon unsafe.Pointer) C.int {
	return C.int(bool2int(mapPolygonContainsPoint((*[2]int32)(unsafe.Pointer(point)), (*mapPolygonRecord)(polygon))))
}

//export nox_xxx_polygonIsPlayerInPolygon_4217B0
func nox_xxx_polygonIsPlayerInPolygon_4217B0(point *C.int2, previous C.int) *C.nox_player_polygon_check_data {
	polygon := findMapPolygonAt((*[2]int32)(unsafe.Pointer(point)), uint32(previous), false)
	return (*C.nox_player_polygon_check_data)(unsafe.Pointer(polygon))
}

//export nox_xxx_polygonGetIdxA_421790
func nox_xxx_polygonGetIdxA_421790(point *C.int2, previous C.int) C.int {
	polygon := findMapPolygonAt((*[2]int32)(unsafe.Pointer(point)), uint32(previous), false)
	if polygon == nil {
		return 0
	}
	return C.int(polygon.Index)
}

//export sub_421880
func sub_421880(point *C.int2, polygon unsafe.Pointer, radius C.float) C.int {
	return C.int(bool2int(mapPolygonNearPoint(point, (*mapPolygonRecord)(polygon), radius)))
}

//export sub_421990
func sub_421990(point *C.int2, radius C.float, previous C.int) *C.int {
	polygon := findMapPolygon(uint32(previous), func(polygon *mapPolygonRecord) bool {
		return mapPolygonNearPoint(point, polygon, radius)
	})
	return (*C.int)(unsafe.Pointer(polygon))
}

//export sub_421F10
func sub_421F10(point *C.int, previous C.int) *C.uchar {
	polygon := findMapPolygonAt((*[2]int32)(unsafe.Pointer(point)), uint32(previous), true)
	return (*C.uchar)(unsafe.Pointer(polygon))
}

//export sub_421B10
func sub_421B10() *C.uint32_t {
	sub_421430()
	sub_421010()
	result := sub_420C70()
	*memmap.PtrUint32(0x5D4594, 588076) = 1
	*memmap.PtrUint16(0x5D4594, polygonDraftCount) = 0
	return result
}

//export sub_421B40
func sub_421B40(index *C.uint32_t) C.int {
	count := memmap.Uint16(0x5D4594, polygonDraftCount)
	for i := range count {
		if memmap.Uint32(0x5D4594, polygonDraftVertices+4*uintptr(i)) == uint32(*index) {
			return 1
		}
	}
	return 0
}
