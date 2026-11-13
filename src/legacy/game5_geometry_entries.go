package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func game5WallCollision(obj *server.Object, normal types.Pointf) {
	Nox_xxx_collSysAddCollision_548630(obj, 0, normal)
}

//export sub_550580
func sub_550580(grid *C.int2, raw *nox_object_t) C.int {
	return C.int(bool2int(GetServer().S().BoxWallAt550580(image.Pt(int(grid.field_0), int(grid.field_4)), asObjectS(raw), collisionSpringNative5488B0, game5WallCollision)))
}

func game5BoxWallSegment(first bool, raw *nox_object_t, current, previous *C.float2, bounds *C.float4, line *C.float2, length C.float) C.int {
	return C.int(bool2int(server.BoxWallSegment550760(first, asObjectS(raw), (*types.Pointf)(unsafe.Pointer(current)), (*types.Pointf)(unsafe.Pointer(previous)), (*[4]float32)(unsafe.Pointer(bounds)), *(*types.Pointf)(unsafe.Pointer(line)), float32(length), collisionSpringNative5488B0, game5WallCollision)))
}

//export sub_550760
func sub_550760(raw *nox_object_t, current, previous *C.float2, bounds *C.float4, line *C.float2, length C.float) C.int {
	return game5BoxWallSegment(false, raw, current, previous, bounds, line, length)
}

//export sub_550A10
func sub_550A10(raw *nox_object_t, current, previous *C.float2, bounds *C.float4, line *C.float2, length C.float) C.int {
	return game5BoxWallSegment(true, raw, current, previous, bounds, line, length)
}

//export sub_550CB0
func sub_550CB0(point, wall *C.float2) C.char {
	return C.char(server.BoxWallQuadrant550CB0(*(*types.Pointf)(unsafe.Pointer(point)), *(*types.Pointf)(unsafe.Pointer(wall))))
}

//export sub_5516A0
func sub_5516A0(segment, rect *C.float4, out *C.float2, limit, allowNegative C.int) C.int {
	if limit <= 0 {
		return 0
	}
	points := server.DoorBoxIntersections5516A0(*(*[4]float32)(unsafe.Pointer(segment)), *(*[4]float32)(unsafe.Pointer(rect)), int(limit), allowNegative != 0)
	if len(points) != 0 {
		copy(unsafe.Slice((*types.Pointf)(unsafe.Pointer(out)), len(points)), points)
	}
	return C.int(len(points))
}

//export sub_551780
func sub_551780(segment *C.float4, x1, x2, y C.float, out *C.float2, allowNegative C.int) C.int {
	point, ok := server.DoorBoxAxis551780(false, *(*[4]float32)(unsafe.Pointer(segment)), float32(x1), float32(x2), float32(y), allowNegative != 0)
	if ok {
		*(*types.Pointf)(unsafe.Pointer(out)) = point
	}
	return C.int(bool2int(ok))
}

//export sub_551870
func sub_551870(segment *C.float4, x, y1, y2 C.float, out *C.float2, allowNegative C.int) C.int {
	point, ok := server.DoorBoxAxis551780(true, *(*[4]float32)(unsafe.Pointer(segment)), float32(y1), float32(y2), float32(x), allowNegative != 0)
	if ok {
		*(*types.Pointf)(unsafe.Pointer(out)) = point
	}
	return C.int(bool2int(ok))
}

//export sub_551960
func sub_551960(segment, segmentBounds, boxBounds *C.float4, out *C.float2) C.int {
	point, ok := server.DoorBoxMidpoint551960(*(*[4]float32)(unsafe.Pointer(segment)), *(*[4]float32)(unsafe.Pointer(segmentBounds)), *(*[4]float32)(unsafe.Pointer(boxBounds)))
	if ok {
		*(*types.Pointf)(unsafe.Pointer(out)) = point
	}
	return C.int(bool2int(ok))
}

//export sub_551A90
func sub_551A90(point *C.float2, rect *C.float4) C.int {
	return C.int(bool2int(server.DoorBoxPoint551A90(*(*types.Pointf)(unsafe.Pointer(point)), *(*[4]float32)(unsafe.Pointer(rect)))))
}
