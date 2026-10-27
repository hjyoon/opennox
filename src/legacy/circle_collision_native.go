package legacy

/*
#include "GAME5.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func circleCollisionDirection550D00(box bool) types.Pointf {
	direction := uintptr(nox_common_randomInt_415FA0(0, 3))
	offset := uintptr(292784)
	if box {
		offset = 289928
	}
	return types.Ptf(memmap.Float32(0x587000, offset+8*direction), memmap.Float32(0x587000, offset+8*direction+4))
}

func circleCollisionRuntimeNative550D00() server.CircleCollisionRuntime550D00 {
	return server.CircleCollisionRuntime550D00{
		PushScale: collisionPushScaleNative550F80, RandomDirection: circleCollisionDirection550D00,
		Trace: func(first, second types.Pointf) bool {
			ray := C.float4{field_0: C.float(first.X), field_4: C.float(first.Y), field_8: C.float(second.X), field_C: C.float(second.Y)}
			return C.nox_xxx_mapTraceRay_535250(&ray, nil, nil, 0) != 0
		},
		AddCollision: func(first, second *server.Object, normal types.Pointf) {
			Nox_xxx_collSysAddCollision_548630(first, uintptr(second.CObj()), normal)
		},
		Wake: func(obj *server.Object) { C.nox_xxx_unitHasCollideOrUpdateFn_537610(asObjectC(obj)) },
	}
}

//export sub_54A990_go
func sub_54A990_go(center *C.float2, radius C.float, box *nox_object_t, normal *C.float2) C.double {
	return C.double(server.CircleBoxGeometry54A990(*(*types.Pointf)(unsafe.Pointer(center)), float32(radius), asObjectS(box), (*types.Pointf)(unsafe.Pointer(normal)), func() types.Pointf { return circleCollisionDirection550D00(true) }))
}

//export sub_54AD50_go
func sub_54AD50_go(circle, box *nox_object_t, move C.int) {
	server.CircleBoxCollision54AD50(asObjectS(circle), asObjectS(box), int32(move), circleCollisionRuntimeNative550D00())
}

//export nox_xxx_collisionCheckCircleCircle_550D00_go
func nox_xxx_collisionCheckCircleCircle_550D00_go(first, second *nox_object_t) {
	server.CircleCircleCollision550D00(asObjectS(first), asObjectS(second), circleCollisionRuntimeNative550D00())
}

func circleWallDefinitionNative54FFC0(index byte) server.CircleWallDefinition54FFC0 {
	d := &C.table_313272[index]
	return server.CircleWallDefinition54FFC0{First: byte(d.f0), Second: byte(d.f12), FirstLo: float32(d.f4), FirstHi: float32(d.f8), SecondLo: float32(d.f16), SecondHi: float32(d.f20)}
}

func circleWallRuntimeNative54FEF0() server.CircleWallRuntime54FEF0 {
	return server.CircleWallRuntime54FEF0{
		Spring: collisionSpringNative5488B0, Definition: circleWallDefinitionNative54FFC0,
		EdgeMask:     func(index byte) byte { return memmap.Uint8(0x587000, 292496+uintptr(index)) },
		WallType:     func(pos image.Point, flags byte) int8 { return GetServer().S().Sub_57B500(pos, flags) },
		WallObject:   func(pos image.Point) *server.Wall { return GetServer().S().Walls.GetWallAtGrid(pos) },
		IsGameBall:   func(obj *server.Object) bool { return C.sub_550480(asObjectC(obj)) != 0 },
		AddCollision: func(obj *server.Object, normal types.Pointf) { Nox_xxx_collSysAddCollision_548630(obj, 0, normal) },
		OpenSecret: func(pos image.Point, obj *server.Object) {
			grid := C.int2{field_0: C.int(pos.X), field_4: C.int(pos.Y)}
			C.sub_548100(&grid, asObjectC(obj))
		},
	}
}

//export sub_54FEF0_go
func sub_54FEF0_go(obj *nox_object_t) {
	server.CircleWallCollision54FEF0(asObjectS(obj), circleWallRuntimeNative54FEF0())
}

//export sub_54FFC0_go
func sub_54FFC0_go(grid *C.int2, obj *nox_object_t) C.int {
	return C.int(bool2int(server.CircleWallAt54FFC0(image.Pt(int(grid.field_0), int(grid.field_4)), asObjectS(obj), circleWallRuntimeNative54FEF0())))
}

//export sub_550380_go
func sub_550380_go(_ C.int, obj *nox_object_t, point *C.float2) C.int {
	return C.int(bool2int(server.CircleWallResponse550380(asObjectS(obj), *(*types.Pointf)(unsafe.Pointer(point)), circleWallRuntimeNative54FEF0())))
}

func circleWallNearestNative550280(second bool, origin *C.float2, lo, hi C.float, allowLo, allowHi C.int, point, result *C.float2) C.int {
	return C.int(bool2int(server.CircleWallNearest550280(second, *(*types.Pointf)(unsafe.Pointer(origin)), float32(lo), float32(hi), int32(allowLo), int32(allowHi), *(*types.Pointf)(unsafe.Pointer(point)), (*types.Pointf)(unsafe.Pointer(result)))))
}

//export sub_550280_go
func sub_550280_go(origin *C.float2, lo, hi C.float, allowLo, allowHi C.int, point, result *C.float2) C.int {
	return circleWallNearestNative550280(false, origin, lo, hi, allowLo, allowHi, point, result)
}

//export sub_5502F0_go
func sub_5502F0_go(origin *C.float2, lo, hi C.float, allowLo, allowHi C.int, point, result *C.float2) C.int {
	return circleWallNearestNative550280(true, origin, lo, hi, allowLo, allowHi, point, result)
}
