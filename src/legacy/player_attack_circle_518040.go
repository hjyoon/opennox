package legacy

/*
#include "GAME4_1.h"

static inline void nox_call_unit_radius_callback(nox_unit_radius_callback_t callback, nox_object_t* candidate, void* data) {
	callback(candidate, data);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerAttackScanCircleNative_518040
func nox_xxx_playerAttackScanCircleNative_518040(center *C.float2, radius C.float, callback, data unsafe.Pointer) {
	sub_518040(center, radius, C.nox_unit_radius_callback_t(callback), data)
}

//export sub_518040
func sub_518040(center *C.float2, radius C.float, callback C.nox_unit_radius_callback_t, data unsafe.Pointer) *C.float2 {
	if center == nil || callback == nil {
		return center
	}
	s := GetServer().S()
	server.PlayerAttackScanCircle518040((*types.Pointf)(unsafe.Pointer(center)), float32(radius),
		s.Map.EachObjInRect, circleBoxCollisionNative54A990,
		func(candidate *server.Object) {
			C.nox_call_unit_radius_callback(callback, asObjectC(candidate), data)
		})
	return center
}

//export sub_5180B0
func sub_5180B0(candidate *C.nox_object_t, context *C.nox_unit_radius_context_t) {
	if context == nil || context.callback == nil {
		return
	}
	center := (*types.Pointf)(unsafe.Pointer(context.center))
	if server.PlayerAttackCircleOverlaps5180B0(center, float32(context.radius), asObjectS(candidate), circleBoxCollisionNative54A990) {
		C.nox_call_unit_radius_callback(context.callback, candidate, context.data)
	}
}
