package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "random_reachable_point_4ed970.h"
// Carry the original nil-object field address into the consuming C entry.
static void debris_point_address(float radius,uintptr_t origin,float2* output) {
 sub_4ED970(radius,(float2*)origin,output);
}
static void debris_force_address(uintptr_t origin,nox_object_t* object,float force) {
 nox_xxx_objectApplyForce_52DF80((float*)origin,object,force);
}
static uint32_t debris_pointfx_address(char id,uintptr_t origin) {
 return (uint32_t)nox_xxx_netSendPointFx_522FF0(id,(float2*)origin);
}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func debrisPoint(radius float32, origin server.SkeletonOrigin54A310, output *types.Pointf) {
	if point := origin.NativePoint(); point != nil {
		GetServer().S().RandomReachablePointAroundInto4ED970(radius, point, output)
		return
	}
	var pin runtime.Pinner
	pin.Pin(output)
	defer pin.Unpin()
	C.debris_point_address(C.float(radius), C.uintptr_t(origin.Address()), (*C.float2)(unsafe.Pointer(output)))
}

func debrisForce(origin server.SkeletonOrigin54A310, o *server.Object, force float32) uint32 {
	if point := origin.NativePoint(); point != nil {
		applyForceExport52DF80(o, unsafe.Pointer(point), force)
	} else {
		C.debris_force_address(C.uintptr_t(origin.Address()), asObjectC(o), C.float(force))
	}
	return 0
}

func debrisPointFX(id uint32, origin server.SkeletonOrigin54A310) uint32 {
	if point := origin.NativePoint(); point != nil {
		return uint32(server.NetSendPointFXNative522FF0(byte(id), point, pointFXRuntimeForCall522FF0()))
	}
	return debrisPointFXAtAddress(id, origin.Address())
}

func debrisPointFXAtAddress(id uint32, address uintptr) uint32 {
	return uint32(C.debris_pointfx_address(C.char(id), C.uintptr_t(address)))
}

// GolemDeathIndex549E90 is the actual extracted C cache shared by ordinary
// Golem callbacks. The old blob address is separate, retired scalar storage.
func GolemDeathIndex549E90() *uint32 {
	return (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2491580))
}
func golemDeathState549E90() *server.GolemDeathState549E90 {
	if golemOriginalStateOverride549E90 != nil {
		return golemOriginalStateOverride549E90
	}
	return &server.GolemDeathState549E90{Index: GolemDeathIndex549E90(), Divisor: memmap.PtrUint32(0x587000, 287344)}
}
func golemDeathRuntime549E90() server.GolemDeathRuntime549E90 {
	if golemOriginalRuntimeOverride549E90 != nil {
		return *golemOriginalRuntimeOverride549E90
	}
	return server.GolemDeathRuntime549E90{
		RandomInt: func(lo, hi int32, _ server.GolemRandomSite549E90) int32 {
			return nox_common_randomInt_415FA0(lo, hi)
		},
		RandomFloat: func(lo, hi float32) float64 {
			return GetServer().S().RandomFloat416030(lo, hi)
		},
		Flags: func() uint32 {
			if noxflags.HasGame(noxflags.GameModeCoop) {
				return 1
			}
			return 0
		},
		FPS: gameFPS,
		MechName: func(slot int32) unsafe.Pointer {
			off := uint32(287976) + uint32(slot)*4
			return *memmap.PtrPtr(0x587000, uintptr(off))
		},
		RockName: func(index int32) unsafe.Pointer {
			off := uint32(288240) + uint32(index)*4
			return *memmap.PtrPtr(0x587000, uintptr(off))
		},
		Mass: func(index int32) byte {
			off := uint32(287332) + uint32(index)
			return *memmap.PtrUint8(0x587000, uintptr(off))
		},
		New: func(name unsafe.Pointer) *server.Object {
			return asObjectS(nox_xxx_newObjectByTypeID_4E3810((*C.char)(name)))
		},
		Point: debrisPoint,
		Create: func(o, owner *server.Object, pos types.Pointf, _ uint32) uint32 {
			nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(pos.X), C.float(pos.Y))
			return 0
		},
		Raise: func(o *server.Object, height float32) uint32 {
			o.Raise(height)
			return 0
		},
		Force: debrisForce,
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			nox_xxx_aud_501960(int32(id), asObjectC(o), int32(kind), int32(arg))
			return 0
		},
		PointFX: debrisPointFX,
		Decay: func(o *server.Object, duration uint32) uint32 {
			return uint32(GetServer().S().DecaySetTime511660(o, duration))
		},
	}
}

//export sub_549E90
func sub_549E90(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).GolemDeathNative549E90(true, golemDeathState549E90(), golemDeathRuntime549E90()))
}

//export sub_549FA0
func sub_549FA0(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).GolemDeathNative549E90(false, golemDeathState549E90(), golemDeathRuntime549E90()))
}
