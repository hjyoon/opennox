package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

func mapgenTileFillOffset(value float32, delta float64) float32 {
	return storeFloat32TowardZero(addFloat64TowardZero(float64(value), delta))
}

func mapgenFillTileRectangle(origin *mapgenWorldPoint, columns, rows int32, place func(*mapgenWorldPoint) uintptr) uintptr {
	if origin == nil {
		return 0
	}
	result := uintptr(int64(rows))
	if columns <= 0 || rows <= 0 {
		return result
	}
	point := mapgenWorldPoint{y: mapgenTileFillOffset(origin.y, mapgenCellSize/2)}
	for row := int32(0); row < rows; row++ {
		point.x = mapgenTileFillOffset(origin.x, mapgenCellSize/2)
		for col := int32(0); col < columns; col++ {
			result = place(&point)
			point.x = mapgenTileFillOffset(point.x, mapgenCellSize)
		}
		point.y = mapgenTileFillOffset(point.y, mapgenCellSize)
	}
	return result
}

func mapgenFillTileColumn(origin *mapgenWorldPoint, rows int32, place func(*mapgenWorldPoint) uintptr) uintptr {
	if origin == nil || rows <= 0 {
		return 0
	}
	point := mapgenWorldPoint{
		x: mapgenTileFillOffset(origin.x, mapgenCellSize/2),
		y: mapgenTileFillOffset(origin.y, mapgenCellSize/2),
	}
	var result uintptr
	for i := int32(0); i < rows; i++ {
		result = place(&point)
		point.y = mapgenTileFillOffset(point.y, mapgenCellSize)
	}
	return result
}

func mapgenPlaceFillTile(position *mapgenWorldPoint) uintptr {
	return uintptr(sub_51D5E0((*C.float2)(unsafe.Pointer(position))))
}

//export sub_5245A0
func sub_5245A0(_ C.int, origin *C.float, columns, rows C.int) C.uintptr_t {
	position := (*mapgenWorldPoint)(unsafe.Pointer(origin))
	return C.uintptr_t(mapgenFillTileRectangle(position, int32(columns), int32(rows), mapgenPlaceFillTile))
}

//export sub_524610
func sub_524610(_ C.int, origin *C.float, rows C.int) C.uintptr_t {
	if rows <= 0 {
		return C.uintptr_t(uintptr(unsafe.Pointer(origin)))
	}
	position := (*mapgenWorldPoint)(unsafe.Pointer(origin))
	return C.uintptr_t(mapgenFillTileColumn(position, int32(rows), mapgenPlaceFillTile))
}
