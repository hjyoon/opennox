package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
static double skull_raw_balance_54f9a0(void){return nox_xxx_gamedataGetFloat_419D40("ArrowTrapDamage");}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func skullRawState54F9A0() *server.SkullRawState54F9A0 {
	if skullRawOriginalStateOverride54F9A0 != nil {
		return skullRawOriginalStateOverride54F9A0
	}
	return (*server.SkullRawState54F9A0)(unsafe.Pointer(memmap.PtrUint32(0x5D4594, 2491764)))
}

// Resolve each world service only when reached by the original raw body.
func skullRawRuntime54F9A0() server.SkullRawRuntime54F9A0 {
	if skullRawOriginalRuntimeOverride54F9A0 != nil {
		return *skullRawOriginalRuntimeOverride54F9A0
	}
	return server.SkullRawRuntime54F9A0{
		Resolve: func(name string) uint32 { return uint32(C.nox_xxx_getNameId_4E3AA0(internCStr(name))) },
		New: func(id uint32) *server.Object {
			return asObjectS(C.nox_xxx_newObjectWithTypeInd_4E3450(C.int(int32(id))))
		},
		Create: func(projectile, owner *server.Object, position types.Pointf, extra uint32) {
			C.nox_xxx_createAt_4DAA50(asObjectC(projectile), asObjectC(owner), C.float(position.X), C.float(position.Y))
		},
		Balance: func() float64 { return float64(C.skull_raw_balance_54f9a0()) },
		Audio: func(id uint32, source *server.Object) {
			C.nox_xxx_aud_501960(C.int(int32(id)), asObjectC(source), 0, 0)
		},
		TrapFX: func(position *types.Pointf, variant uint32) {
			C.nox_xxx_sendArrowTrapFX_5238A0((*C.float)(unsafe.Pointer(position)), C.char(variant))
		},
		Enumerate: func(rect types.Rectf, source *server.Object, callback func(*server.Object, *server.Object)) {
			GetServer().S().Map.EachObjInRect(rect, func(candidate *server.Object) bool { callback(candidate, source); return true })
		},
		Enemy: func(source, candidate *server.Object) uint32 {
			return uint32(C.nox_xxx_unitIsEnemyTo_5330C0(asObjectC(source), asObjectC(candidate)))
		},
		MapCheck: func(candidate, source *server.Object) uint32 {
			return uint32(C.nox_xxx_mapCheck_537110(asObjectC(candidate), asObjectC(source)))
		},
		Front: func(source *server.Object, direction int32, candidate *server.Object) uint32 {
			return uint32(C.nox_server_testTwoPointsAndDirection_4E6E50((*C.float2)(unsafe.Pointer(&source.PosVec)), C.int(direction), (*C.float2)(unsafe.Pointer(&candidate.PosVec))))
		},
		FPS: func() uint32 { return uint32(C.gameFPS()) },
		DirectionComponent: func(direction int16, component uint32) float32 {
			return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(direction)*8 + int64(component)*4)))
		},
	}
}

//export nox_xxx_updateShootingTrap_54F9A0
func nox_xxx_updateShootingTrap_54F9A0(source *nox_object_t) C.int {
	return C.int(asObjectS(source).SkullRawUpdate54F9A0(skullRawState54F9A0(), skullRawRuntime54F9A0()))
}

//export nox_xxx_createArrowTrapProjectile_54FA80
func nox_xxx_createArrowTrapProjectile_54FA80(source *nox_object_t, id C.int) {
	asObjectS(source).SkullRawProjectile54FA80(uint32(id), skullRawState54F9A0(), skullRawRuntime54F9A0())
}

//export sub_54FBB0
func sub_54FBB0(source *nox_object_t) {
	asObjectS(source).SkullRawState54FBB0(skullRawState54F9A0(), skullRawRuntime54F9A0())
}

//export sub_54FBF0
func sub_54FBF0(source *nox_object_t) C.int {
	return C.int(asObjectS(source).SkullRawScan54FBF0(skullRawState54F9A0(), skullRawRuntime54F9A0()))
}

//export nox_xxx_unitIsAttackReachable_54FC50
func nox_xxx_unitIsAttackReachable_54FC50(candidate, source *nox_object_t) {
	asObjectS(source).SkullRawReachable54FC50(asObjectS(candidate), skullRawState54F9A0(), skullRawRuntime54F9A0())
}
