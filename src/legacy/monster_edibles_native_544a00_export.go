package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
#include "GAME5.h"
_Static_assert(_Generic(&nox_xxx_mobRetreatCheckEdibles_5455E0,void* (*)(nox_object_t*):1,default:0), "native edible parent object and pointer-return ABI");
_Static_assert(_Generic(&nox_xxx_mobSearchEdible_544A00,nox_object_t* (*)(nox_object_t*,float):1,default:0), "native edible search object and pointer-return ABI");
_Static_assert(_Generic(&nox_xxx_mobSearchEdible2_544A40,void (*)(nox_object_t*,nox_object_t*):1,default:0), "native edible callback object ABI");
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func edibleNativeRuntime544A00() server.MonsterEdibleNativeRuntime544A00 {
	if edibleNativeRuntimeOverride544A00 != nil {
		return *edibleNativeRuntimeOverride544A00
	}
	s := GetServer().S()
	return server.MonsterEdibleNativeRuntime544A00{
		EachCircle: func(point *types.Pointf, radius float32, callback func(*server.Object, *server.Object), unit *server.Object) {
			s.Map.EachObjInCircle(*point, radius, func(candidate *server.Object) bool { callback(candidate, unit); return true })
		},
		CanInteract: s.CanInteract,
		Quest:       func() bool { return edibleOriginalNativeFlags544A00(4096) },
		Online:      func() bool { return edibleOriginalNativeFlags544A00(8192) },
		Push: func(unit *server.Object, action ai.ActionType) *server.AIStackItem {
			return unit.MonsterPushAction(action)
		},
	}
}

//export nox_xxx_mobRetreatCheckEdibles_5455E0
func nox_xxx_mobRetreatCheckEdibles_5455E0(unit *C.nox_object_t) unsafe.Pointer {
	return unsafe.Pointer((*server.Object)(unsafe.Pointer(unit)).MonsterRetreatCheckEdiblesNative5455E0(&edibleNativeScratch544A00, edibleNativeRuntime544A00()))
}

//export nox_xxx_mobSearchEdible_544A00
func nox_xxx_mobSearchEdible_544A00(unit *C.nox_object_t, radius C.float) *C.nox_object_t {
	return (*C.nox_object_t)(unsafe.Pointer((*server.Object)(unsafe.Pointer(unit)).MonsterSearchEdibleNative544A00(float32(radius), &edibleNativeScratch544A00, edibleNativeRuntime544A00())))
}

//export nox_xxx_mobSearchEdible2_544A40
func nox_xxx_mobSearchEdible2_544A40(candidate, unit *C.nox_object_t) {
	(*server.Object)(unsafe.Pointer(candidate)).MonsterSearchEdibleCandidateNative544A40((*server.Object)(unsafe.Pointer(unit)), &edibleNativeScratch544A00, edibleNativeRuntime544A00())
}
