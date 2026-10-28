package legacy

/*
#include "GAME1.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "unit_buff_test_4ff350.h"
void nox_xxx_updateSprings_5113A0(void);
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var objectsVelocityRuntime5118A0 = defaultObjectsVelocityRuntime5118A0

func defaultObjectsVelocityRuntime5118A0() server.ObjectsVelocityRuntime5118A0 {
	return server.ObjectsVelocityRuntime5118A0{
		Cache:   (*[10]uint32)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 2386580))),
		Lookup:  func(name string) int32 { return int32(GetServer().S().Types.IndByID(name)) },
		First:   func() *server.Object { return asObjectS(C.sub_537740()) },
		Next:    func(o *server.Object) *server.Object { return asObjectS(C.sub_537750(asObjectC(o))) },
		Prepare: func(o *server.Object) { C.sub_5481C0(asObjectC(o)) },
		Springs: func() { C.nox_xxx_updateSprings_5113A0() },
		Action: func(o *server.Object, action int32) int32 {
			return int32(C.nox_xxx_checkMobAction_50A0D0(asObjectC(o), C.int(action)))
		},
		Buff: func(o *server.Object, buff int32) int32 {
			return int32(C.nox_xxx_testUnitBuffs_4FF350(asObjectC(o), C.int32_t(buff)))
		},
		Trace: func(ray *types.Rectf, flags byte) int32 {
			return int32(C.nox_xxx_mapTraceRay_535250((*C.float4)(unsafe.Pointer(ray)), nil, nil, C.char(flags)))
		},
		Tile: func(point *types.Pointf) int32 {
			return int32(C.nox_xxx_tileNFromPoint_411160((*C.float2)(unsafe.Pointer(point))))
		},
		Collision: func(o *server.Object, kind uintptr, normal *types.Pointf) {
			C.nox_xxx_collSysAddCollision_548630(asObjectC(o), C.uintptr_t(kind), (*C.float2)(unsafe.Pointer(normal)))
		},
		Sync:   func(o *server.Object) { C.nox_xxx_unitNeedSync_4E44F0(asObjectC(o)) },
		Coords: func(o *server.Object) { C.nox_xxx_objectUnkUpdateCoords_4E7290(asObjectC(o)) },
		Move:   func(o *server.Object) { C.nox_xxx_moveUpdateSpecial_517970(asObjectC(o)) },
	}
}

func Nox_xxx_updateObjectsVelocity_5118A0(step float32) {
	server.UpdateObjectsVelocity5118A0(step, objectsVelocityRuntime5118A0())
}

//export nox_update_objects_velocity_bits_5118a0
func nox_update_objects_velocity_bits_5118a0(step C.uint32_t) C.int32_t {
	return C.int32_t(server.UpdateObjectsVelocity5118A0(math.Float32frombits(uint32(step)), objectsVelocityRuntime5118A0()))
}
