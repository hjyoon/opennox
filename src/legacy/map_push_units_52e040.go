package legacy

/*
#include "map_push_units_52e040.h"

static inline void nox_map_push_before_52E0E0(uintptr_t callback, nox_object_t* candidate, uint32_t distance, uintptr_t arg) {
	if (callback) ((void (*)(nox_object_t*, uint32_t, uintptr_t))callback)(candidate, distance, arg);
}
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func mapPushRuntime52E040(callback, arg int) server.MapPushUnitsAroundRuntime52E040 {
	runtime := server.MapPushUnitsAroundRuntime52E040{Wake: Nox_xxx_unitHasCollideOrUpdateFn_537610}
	if callback != 0 {
		runtime.BeforeForce = func(p *server.Object, distance float32) {
			C.nox_map_push_before_52E0E0(C.uintptr_t(uintptr(callback)), asObjectC(p), C.uint32_t(math.Float32bits(distance)), C.uintptr_t(uintptr(arg)))
		}
	}
	return runtime
}

func mapPushUnitsAroundNative52E040(pos *types.Pointf, outer, inner, force float32, _ *server.Object, callback, arg int) {
	GetServer().S().MapPushUnitsAroundPtr52E040(pos, outer, inner, force, mapPushRuntime52E040(callback, arg))
}

// Retain the dynamic native dispatch used by the original public entry.
var mapPushUnitsAroundCall52E040 = mapPushUnitsAroundNative52E040

func Nox_xxx_mapPushUnitsAround_52E040(pos types.Pointf, outer, inner, force float32, source *server.Object, callback, arg int) {
	mapPushUnitsAroundCall52E040(&pos, outer, inner, force, source, callback, arg)
}

func Nox_xxx_mapPushUnitsAroundPtr52E040(pos *types.Pointf, outer, inner, force float32, source *server.Object, callback, arg int) {
	mapPushUnitsAroundCall52E040(pos, outer, inner, force, source, callback, arg)
}

//export nox_xxx_mapPushUnitsAround_native_52E040
func nox_xxx_mapPushUnitsAround_native_52E040(pos unsafe.Pointer, outer, inner, force C.float, source *nox_object_t, callback, arg C.uintptr_t) {
	if pos == nil {
		return
	}
	mapPushUnitsAroundCall52E040((*types.Pointf)(pos), float32(outer), float32(inner), float32(force), asObjectS(source), int(uintptr(callback)), int(uintptr(arg)))
}

//export nox_xxx_unitPushAroundFn_native_52E0E0
func nox_xxx_unitPushAroundFn_native_52E0E0(candidate *nox_object_t, args *C.nox_map_push_args_52E0E0) {
	if candidate == nil || args == nil || args.origin == nil {
		return
	}
	GetServer().S().MapPushUnit52E0E0(asObjectS(candidate), server.MapPushUnitArguments52E0E0{
		Origin: func() types.Pointf { return *(*types.Pointf)(unsafe.Pointer(args.origin)) },
		Radius: func() float32 { return float32(args.radius) }, Inner: func() float32 { return float32(args.inner) }, ScaledForce: func() float32 { return float32(args.scaled_force) },
	}, server.MapPushUnitsAroundRuntime52E040{
		BeforeForce: func(p *server.Object, distance float32) {
			C.nox_map_push_before_52E0E0(args.callback, asObjectC(p), C.uint32_t(math.Float32bits(distance)), args.callback_arg)
		},
		Wake: Nox_xxx_unitHasCollideOrUpdateFn_537610,
	})
}

func mapPushUnitsAroundLegacyEntry52E040(pos types.Pointf, outer, inner, force float32, source *server.Object, callback, arg int) {
	C.nox_xxx_mapPushUnitsAround_52E040(unsafe.Pointer(&pos), C.float(outer), C.float(inner), C.float(force), asObjectC(source), C.uintptr_t(uintptr(callback)), C.uintptr_t(uintptr(arg)))
}
