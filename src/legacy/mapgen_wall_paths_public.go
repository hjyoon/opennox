package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

const mapgenWallCoordinateScale = 0x1.642c86p-5

func mapgenWallGridCell(x, y float32) (int32, int32, bool) {
	col := x87TruncQwordLow(mulFloat64TowardZero(float64(x), mapgenWallCoordinateScale))
	row := x87TruncQwordLow(mulFloat64TowardZero(float64(y), mapgenWallCoordinateScale))
	valid := col >= 0 && col < 256 && row >= 0 && row < 256 && (col+row)&1 == 0
	return col, row, valid
}

func mapgenWalkWallStrip(point [2]float32, count int32, axis int, visit func([2]float32)) {
	for i := int32(0); i < count; i++ {
		visit(point)
		point[axis] = storeFloat32TowardZero(addFloat64TowardZero(float64(point[axis]), mapgenCellSize))
	}
}

func mapgenRemoveWallStrip(pos *C.float2, count C.int, axis int) {
	point := [2]float32{float32(pos.field_0), float32(pos.field_4)}
	mapgenWalkWallStrip(point, int32(count), axis, func(at [2]float32) {
		visit := C.float2{field_0: C.float(at[0]), field_4: C.float(at[1])}
		sub_527030(&visit)
	})
}

//export sub_525330
func sub_525330(pos *C.float2, count C.int) {
	mapgenRemoveWallStrip(pos, count, 0)
}

//export sub_525370
func sub_525370(pos *C.float2, count C.int) {
	mapgenRemoveWallStrip(pos, count, 1)
}

//export sub_526DD0
func sub_526DD0(pos *C.float2, out *unsafe.Pointer) C.int {
	var fixed C.float2
	if C.nox_xxx_mapGenFixCoords_4D3D90(pos, &fixed) == 0 || out == nil {
		return 0
	}
	x, y, valid := mapgenWallGridCell(float32(fixed.field_0), float32(fixed.field_4))
	if !valid {
		return 0
	}
	wall := nox_server_getWallAtGrid_410580(x, y)
	if wall == nil {
		return 0
	}
	*out = wall
	return 1
}

//export sub_527030
func sub_527030(pos *C.float2) C.int {
	var fixed C.float2
	if C.nox_xxx_mapGenFixCoords_4D3D90(pos, &fixed) == 0 {
		return 0
	}
	x, y, valid := mapgenWallGridCell(float32(fixed.field_0), float32(fixed.field_4))
	if !valid {
		return 0
	}
	wall := nox_server_getWallAtGrid_410580(x, y)
	if wall == nil {
		return 0
	}
	if data := nox_server_wallData(wall); data != nil {
		sub_4107A0(data)
		nox_server_wallSetData(wall, nil)
	}
	nox_xxx_mapDelWallAtPt_410430(x, y)
	return 1
}
