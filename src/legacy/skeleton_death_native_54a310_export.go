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
int nox_common_randomInt_415FA0(int min, int max);
double nox_common_randomFloat_416030(float min, float max);
extern uint32_t dword_5d4594_2491588;
#include "decay_511660.h"
#include "random_reachable_point_4ed970.h"
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// SkeletonDeathIndex54A310 is the extracted C DWORD shared with the ordinary
// monster callback. Its former blob slot is separate storage.
func SkeletonDeathIndex54A310() *uint32 {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_2491588))
}

func skeletonDeathState54A310() *server.SkeletonDeathState54A310 {
	if skeletonOriginalStateOverride54A310 != nil {
		return skeletonOriginalStateOverride54A310
	}
	return &server.SkeletonDeathState54A310{Index: SkeletonDeathIndex54A310(), Divisor: memmap.PtrUint32(0x587000, 287348)}
}
func skeletonDeathRuntime54A310() server.SkeletonDeathRuntime54A310 {
	if skeletonOriginalRuntimeOverride54A310 != nil {
		return *skeletonOriginalRuntimeOverride54A310
	}
	return server.SkeletonDeathRuntime54A310{
		Names: server.SkeletonDeathNames54A310{Skull: skeletonOriginalName54A310(5), Sword: skeletonOriginalName54A310(8), WoodenShield: skeletonOriginalName54A310(9), SteelShield: skeletonOriginalName54A310(10), Power: skeletonOriginalName54A310(1), Material: skeletonOriginalName54A310(2)},
		RandomInt: func(lo, hi int32, _ server.SkeletonRandomSite54A310) int32 {
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
		BoneName: func(index int32) unsafe.Pointer {
			// Original PE32 scale/add wraps before the packed pointer slot lookup;
			// PtrPtr supplies the native side slot rather than reading a DWORD address.
			off := uint32(288868) + uint32(index)*4
			return *memmap.PtrPtr(0x587000, uintptr(off))
		},
		New: func(name unsafe.Pointer) *server.Object {
			return asObjectS(C.nox_xxx_newObjectByTypeID_4E3810((*C.char)(name)))
		},
		Point: func(radius float32, origin, output *types.Pointf) {
			// This service returns the output pointer through its C export. Keep
			// the Go-owned point pinned until that return crosses back into Go.
			var pin runtime.Pinner
			pin.Pin(output)
			defer pin.Unpin()
			C.sub_4ED970(C.float(radius), (*C.float2)(unsafe.Pointer(origin)), (*C.float2)(unsafe.Pointer(output)))
		},
		Create: func(o, owner *server.Object, pos types.Pointf, _ uint32) uint32 {
			C.nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(pos.X), C.float(pos.Y))
			return 0
		},
		Raise: func(o *server.Object, height float32) uint32 {
			C.nox_xxx_unitRaise_4E46F0(asObjectC(o), C.float(height))
			return 0
		},
		Force: func(origin *types.Pointf, o *server.Object, force float32) uint32 {
			C.nox_xxx_objectApplyForce_52DF80((*C.float)(unsafe.Pointer(origin)), asObjectC(o), C.float(force))
			return 0
		},
		Audio: func(id uint32, o *server.Object, kind, arg uint32) uint32 {
			C.nox_xxx_aud_501960(C.int(id), asObjectC(o), C.int(kind), C.int(arg))
			return 0
		},
		PointFX: func(id uint32, origin *types.Pointf) uint32 {
			return uint32(C.nox_xxx_netSendPointFx_522FF0(C.char(id), (*C.float2)(unsafe.Pointer(origin))))
		},
		Decay: func(o *server.Object, duration uint32) uint32 {
			return uint32(C.nox_xxx_unitSetDecayTime_511660(asObjectC(o), C.int(int32(duration))))
		},
		ModifierID:   func(name unsafe.Pointer) int32 { return int32(C.nox_xxx_modifGetIdByName_413290((*C.char)(name))) },
		ModifierDesc: func(id int32) unsafe.Pointer { return C.nox_xxx_modifGetDescById_413330(C.int(id)) },
		Apply: func(o *server.Object, attrs *server.ModifierInitData) uint32 {
			C.nox_xxx_modifSetItemAttrs_4E4990(asObjectC(o), (*C.nox_modifier_attrs_t)(unsafe.Pointer(attrs)))
			return 0
		},
		ArmorSound:  func(o *server.Object) uint32 { C.sub_53EAE0(asObjectC(o)); return 0 },
		WeaponSound: func(o *server.Object) uint32 { C.sub_53AAB0(asObjectC(o)); return 0 },
	}
}

//export sub_54A310
func sub_54A310(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).SkeletonDeathNative54A310(false, skeletonDeathState54A310(), skeletonDeathRuntime54A310()))
}

//export sub_54A390
func sub_54A390(unit *nox_object_t, name, m0, m1, m2, m3 *C.char, charge C.int) {
	asObjectS(unit).SkeletonDropNative54A390(unsafe.Pointer(name), [4]unsafe.Pointer{unsafe.Pointer(m0), unsafe.Pointer(m1), unsafe.Pointer(m2), unsafe.Pointer(m3)}, int32(charge), skeletonDeathRuntime54A310())
}

//export sub_54A4C0
func sub_54A4C0(unit *nox_object_t) {
	asObjectS(unit).SkeletonDebrisNative54A4C0(skeletonDeathState54A310(), skeletonDeathRuntime54A310())
}

//export sub_54A750
func sub_54A750(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).SkeletonDeathNative54A310(true, skeletonDeathState54A310(), skeletonDeathRuntime54A310()))
}
