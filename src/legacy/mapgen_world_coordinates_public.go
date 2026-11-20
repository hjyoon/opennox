package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

type mapgenWorldPoint struct {
	x, y float32
}

// The original coefficient is binary32, even when it is used by x87
// instructions running with 53-bit precision.
const mapgenWorldRotate = float64(float32(0.70710677))

func clampMapgenWorldPoint(point *mapgenWorldPoint) {
	// The first x87 comparison clamps unordered coordinates as well.
	if !(point.x > 80.5) {
		point.x = 82.5
	}
	if !(point.y > 80.5) {
		point.y = 81.5
	}
	if point.x >= 5853.5 {
		point.x = 5851.5
	}
	if point.y >= 5853.5 {
		point.y = 5852.5
	}
}

func mapgenFixWorldCoords(input, output *mapgenWorldPoint) bool {
	if input == nil || output == nil {
		return false
	}
	sum := addFloat64TowardZero(float64(input.y), float64(input.x))
	worldX := addFloat64TowardZero(mulFloat64TowardZero(sum, mapgenWorldRotate), 2957)
	output.x = storeFloat32TowardZero(worldX)
	// X is stored before the next two loads, including overlapping pairs.
	difference := addFloat64TowardZero(float64(input.y), -float64(input.x))
	worldY := addFloat64TowardZero(mulFloat64TowardZero(difference, mapgenWorldRotate), 2956)
	output.y = storeFloat32TowardZero(worldY)
	clampMapgenWorldPoint(output)
	return true
}

func mapgenInvertWorldCoords(input, output *mapgenWorldPoint) bool {
	if input == nil || output == nil {
		return false
	}
	clampMapgenWorldPoint(input)
	x := addFloat64TowardZero(float64(input.x), -1)
	x = addFloat64TowardZero(x, -float64(input.y))
	output.x = storeFloat32TowardZero(mulFloat64TowardZero(x, mapgenWorldRotate))
	y := addFloat64TowardZero(float64(input.y), float64(input.x))
	y = addFloat64TowardZero(y, -5912)
	output.y = storeFloat32TowardZero(mulFloat64TowardZero(y, mapgenWorldRotate))
	return true
}

//export nox_xxx_mapGenFixCoords_4D3D90
func nox_xxx_mapGenFixCoords_4D3D90(input, output *C.float2) C.int {
	return C.int(bool2int(mapgenFixWorldCoords(
		(*mapgenWorldPoint)(unsafe.Pointer(input)), (*mapgenWorldPoint)(unsafe.Pointer(output)))))
}

//export sub_4D3E30
func sub_4D3E30(input, output *C.float2) C.int {
	return C.int(bool2int(mapgenInvertWorldCoords(
		(*mapgenWorldPoint)(unsafe.Pointer(input)), (*mapgenWorldPoint)(unsafe.Pointer(output)))))
}
