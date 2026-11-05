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
#include "server__script__script.h"
#include "random_reachable_point_4ed970.h"
int nox_common_randomInt_415FA0(int min,int max);
double nox_common_randomFloat_416030(float min,float max);
extern uint32_t dword_5d4594_2491704;
// Only malformed nil origins use scalar addresses across Go. Normal origins
// retain typed native pointers, and the original offset is formed only in C.
static void special_death_point_address_54e460(float radius,uintptr_t origin,float2* output) {
 sub_4ED970(radius,(float2*)origin,output);
}
static void special_death_force_address_54e460(uintptr_t origin,nox_object_t* object,float force) {
 nox_xxx_objectApplyForce_52DF80((float*)origin,object,force);
}
static uint32_t special_death_pointfx_address_54e460(char id,uintptr_t origin) {
 return (uint32_t)nox_xxx_netSendPointFx_522FF0(id,(float2*)origin);
}
static void special_death_script_address_54e460(uintptr_t data,nox_object_t* caller,nox_object_t* trigger,uint32_t event) {
 nox_xxx_scriptCallByEventBlock_502490((void*)data,caller,trigger,(int)event);
}
static void special_death_generator_address_54e460(uintptr_t origin,uint32_t intensity) {
 nox_xxx_sendGeneratorBreakFX_523200((float*)origin,(char)intensity);
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

// BoulderDeathIndex54E4B0 is the actual extracted C cache also shared by the
// ordinary Boulder dispatcher. It is distinct from the Golem debris cache.
func BoulderDeathIndex54E4B0() *uint32 { return (*uint32)(unsafe.Pointer(&C.dword_5d4594_2491704)) }
func specialDeathState54E460() *server.GolemDeathState549E90 {
	if specialDeathOriginalStateOverride54E460 != nil {
		return specialDeathOriginalStateOverride54E460
	}
	return &server.GolemDeathState549E90{Index: BoulderDeathIndex54E4B0(), Divisor: memmap.PtrUint32(0x587000, 290340)}
}
func specialDeathRuntime54E460() server.SpecialDeathRuntime54E460 {
	if specialDeathOriginalRuntimeOverride54E460 != nil {
		return *specialDeathOriginalRuntimeOverride54E460
	}
	return server.SpecialDeathRuntime54E460{GolemDeathRuntime549E90: server.GolemDeathRuntime549E90{
		RandomInt: func(lo, hi int32, _ server.GolemRandomSite549E90) int32 {
			return int32(C.nox_common_randomInt_415FA0(C.int(lo), C.int(hi)))
		},
		RandomFloat: func(lo, hi float32) float64 {
			return float64(C.nox_common_randomFloat_416030(C.float(lo), C.float(hi)))
		},
		FPS: func() uint32 { return uint32(C.gameFPS()) },
		RockName: func(index int32) unsafe.Pointer {
			off := uint32(291512) + uint32(index)*4
			return *memmap.PtrPtr(0x587000, uintptr(off))
		},
		Mass: func(index int32) byte {
			off := uint32(290328) + uint32(index)
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
				C.special_death_point_address_54e460(C.float(radius), C.uintptr_t(origin.Address()), (*C.float2)(unsafe.Pointer(output)))
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
				C.special_death_force_address_54e460(C.uintptr_t(origin.Address()), asObjectC(o), C.float(force))
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
			return uint32(C.special_death_pointfx_address_54e460(C.char(id), C.uintptr_t(origin.Address())))
		},
		Decay: func(o *server.Object, duration uint32) uint32 {
			return uint32(C.nox_xxx_unitSetDecayTime_511660(asObjectC(o), C.int(int32(duration))))
		},
	},
		DestroyedName: unsafe.Pointer(internCStr("DestroyedGenerator")),
		Parent: func(unit *server.Object) *server.Object {
			return asObjectS(C.nox_xxx_findParentChainPlayer_4EC580(asObjectC(unit)))
		},
		Delete: func(unit *server.Object) uint32 { C.nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(unit)); return 0 },
		Frame:  func() uint32 { return uint32(C.gameFrame()) },
		Timer:  func(frame uint32) uint32 { return uint32(C.sub_4D71E0(C.int(int32(frame)))) },
		Mode:   func(mode int32) uint32 { return uint32(C.sub_4D7520(C.int(mode))) },
		Script: func(data server.SpecialDeathCallback54E460, caller, trigger *server.Object, event uint32) uint32 {
			if callback := data.NativeCallback(); callback != nil {
				C.nox_xxx_scriptCallByEventBlock_502490(unsafe.Pointer(callback), unsafe.Pointer(caller), unsafe.Pointer(trigger), C.int(event))
			} else {
				C.special_death_script_address_54e460(C.uintptr_t(data.Address()), asObjectC(caller), asObjectC(trigger), C.uint32_t(event))
			}
			return 0
		},
		GeneratorFX: func(origin server.SkeletonOrigin54A310, intensity uint32) uint32 {
			if point := origin.NativePoint(); point != nil {
				C.nox_xxx_sendGeneratorBreakFX_523200((*C.float)(unsafe.Pointer(point)), C.char(intensity))
			} else {
				C.special_death_generator_address_54e460(C.uintptr_t(origin.Address()), C.uint32_t(intensity))
			}
			return 0
		},
		Quest: func() int32 {
			if bool(C.nox_common_gameFlags_check_40A5C0(4096)) {
				return 1
			}
			return 0
		},
		Credit: func(unit *server.Object) uint32 { C.sub_4D61B0(asObjectC(unit)); return 0 },
	}
}

//export nox_xxx_dieMarker_54E460
func nox_xxx_dieMarker_54E460(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(0, specialDeathState54E460(), specialDeathRuntime54E460())
}

//export nox_xxx_dieBoulder_54E4B0
func nox_xxx_dieBoulder_54E4B0(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(1, specialDeathState54E460(), specialDeathRuntime54E460())
}

//export nox_xxx_dieMonsterGen_54E630
func nox_xxx_dieMonsterGen_54E630(unit *nox_object_t) {
	asObjectS(unit).SpecialDeathNative54E460(2, specialDeathState54E460(), specialDeathRuntime54E460())
}
