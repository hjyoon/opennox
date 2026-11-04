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
#include "decay_511660.h"
#include "random_reachable_point_4ed970.h"
int nox_common_randomInt_415FA0(int min,int max);
double nox_common_randomFloat_416030(float min,float max);
extern uint32_t dword_5d4594_2491580;
// Only malformed nil origins use scalar addresses across Go. Normal origins
// retain typed native pointers, and the original offset is formed only in C.
static void golem_point_address_549e90(float radius,uintptr_t origin,float2* output) {
 sub_4ED970(radius,(float2*)origin,output);
}
static void golem_force_address_549e90(uintptr_t origin,nox_object_t* object,float force) {
 nox_xxx_objectApplyForce_52DF80((float*)origin,object,force);
}
static uint32_t golem_pointfx_address_549e90(char id,uintptr_t origin) {
 return (uint32_t)nox_xxx_netSendPointFx_522FF0(id,(float2*)origin);
}
*/
import "C"
import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
	"runtime"
	"unsafe"
)

// GolemDeathIndex549E90 is the actual extracted C cache shared by ordinary
// Golem callbacks. The old blob address is separate, retired scalar storage.
func GolemDeathIndex549E90() *uint32 { return (*uint32)(unsafe.Pointer(&C.dword_5d4594_2491580)) }
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
			return int32(C.nox_common_randomInt_415FA0(C.int(lo), C.int(hi)))
		},
		RandomFloat: func(lo, hi float32) float64 {
			return float64(C.nox_common_randomFloat_416030(C.float(lo), C.float(hi)))
		},
		Flags: func() uint32 {
			if bool(C.nox_common_gameFlags_check_40A5C0(2048)) {
				return 1
			}
			return 0
		},
		FPS: func() uint32 { return uint32(C.gameFPS()) },
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
			return asObjectS(C.nox_xxx_newObjectByTypeID_4E3810((*C.char)(name)))
		},
		Point: func(radius float32, origin server.GolemOrigin549E90, output *types.Pointf) {
			var pin runtime.Pinner
			pin.Pin(output)
			defer pin.Unpin()
			if point := origin.NativePoint(); point != nil {
				C.sub_4ED970(C.float(radius), (*C.float2)(unsafe.Pointer(point)), (*C.float2)(unsafe.Pointer(output)))
			} else {
				C.golem_point_address_549e90(C.float(radius), C.uintptr_t(origin.Address()), (*C.float2)(unsafe.Pointer(output)))
			}
		},
		Create: func(o, owner *server.Object, pos types.Pointf, _ uint32) uint32 {
			C.nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(pos.X), C.float(pos.Y))
			return 0
		},
		Raise: func(o *server.Object, height float32) uint32 {
			C.nox_xxx_unitRaise_4E46F0(asObjectC(o), C.float(height))
			return 0
		},
		Force: func(origin server.GolemOrigin549E90, o *server.Object, force float32) uint32 {
			if point := origin.NativePoint(); point != nil {
				C.nox_xxx_objectApplyForce_52DF80((*C.float)(unsafe.Pointer(point)), asObjectC(o), C.float(force))
			} else {
				C.golem_force_address_549e90(C.uintptr_t(origin.Address()), asObjectC(o), C.float(force))
			}
			return 0
		},
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			C.nox_xxx_aud_501960(C.int(id), asObjectC(o), C.int(kind), C.int(arg))
			return 0
		},
		PointFX: func(id uint32, origin server.GolemOrigin549E90) uint32 {
			if point := origin.NativePoint(); point != nil {
				return uint32(C.nox_xxx_netSendPointFx_522FF0(C.char(id), (*C.float2)(unsafe.Pointer(point))))
			}
			return uint32(C.golem_pointfx_address_549e90(C.char(id), C.uintptr_t(origin.Address())))
		},
		Decay: func(o *server.Object, duration uint32) uint32 {
			return uint32(C.nox_xxx_unitSetDecayTime_511660(asObjectC(o), C.int(int32(duration))))
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
