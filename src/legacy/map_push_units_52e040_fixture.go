package legacy

/*
#include <stddef.h>
#include "map_push_units_52e040.h"
typedef struct nox_radial_probe_52E0E0 {
	uint32_t token;
	uint32_t calls;
	uintptr_t unit;
	uint32_t distance;
	uintptr_t arg;
	float2* origin;
} nox_radial_probe_52E0E0;
static void nox_radial_probe_callback_52E0E0(nox_object_t* unit, uint32_t distance, uintptr_t arg) {
	nox_radial_probe_52E0E0* p = (nox_radial_probe_52E0E0*)arg;
	p->calls++;
	p->unit = (uintptr_t)unit;
	p->distance = distance;
	p->arg = arg;
	if (p->calls == 1 && p->origin) {
		((float*)p->origin)[0] = 302.25f;
		((float*)p->origin)[1] = 299.75f;
	}
}
static inline uintptr_t nox_radial_probe_pointer_52E0E0(void) { return (uintptr_t)nox_radial_probe_callback_52E0E0; }
enum {
	nox_radial_probe_calls = offsetof(nox_radial_probe_52E0E0,calls),
	nox_radial_probe_unit = offsetof(nox_radial_probe_52E0E0,unit),
	nox_radial_probe_distance = offsetof(nox_radial_probe_52E0E0,distance),
	nox_radial_probe_arg = offsetof(nox_radial_probe_52E0E0,arg)
};
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func radialPushCSize52E0E0() (uintptr, uintptr) {
	return C.sizeof_nox_map_push_args_52E0E0, C.sizeof_nox_radial_probe_52E0E0
}
func radialPushCProbePointer52E0E0() uintptr { return uintptr(C.nox_radial_probe_pointer_52E0E0()) }
func radialPushCProbeOffsets52E0E0() [4]uintptr {
	return [4]uintptr{C.nox_radial_probe_calls, C.nox_radial_probe_unit, C.nox_radial_probe_distance, C.nox_radial_probe_arg}
}
func radialPushCInit52E0E0(data unsafe.Pointer, origin *types.Pointf, radius, inner, scaled float32, source *server.Object, callback, arg uintptr) {
	*(*C.nox_map_push_args_52E0E0)(data) = C.nox_map_push_args_52E0E0{origin: (*C.float2)(unsafe.Pointer(origin)), radius: C.float(radius), inner: C.float(inner), scaled_force: C.float(scaled), source: asObjectC(source), callback: C.uintptr_t(callback), callback_arg: C.uintptr_t(arg)}
}
func radialPushCProbeInit52E0E0(data unsafe.Pointer) {
	*(*C.nox_radial_probe_52E0E0)(data) = C.nox_radial_probe_52E0E0{token: 0x12345678}
}

func radialPushCProbeOrigin52E0E0(data unsafe.Pointer, origin *types.Pointf) {
	(*C.nox_radial_probe_52E0E0)(data).origin = (*C.float2)(unsafe.Pointer(origin))
}
func radialPushCCandidate52E0E0(unit *server.Object, data unsafe.Pointer) {
	C.nox_xxx_unitPushAroundFn_52E0E0(asObjectC(unit), (*C.nox_map_push_args_52E0E0)(data))
}
func radialPushCRoot52E040(origin *types.Pointf, outer, inner, force float32, source *server.Object, callback, arg uintptr) {
	C.nox_xxx_mapPushUnitsAround_52E040(unsafe.Pointer(origin), C.float(outer), C.float(inner), C.float(force), asObjectC(source), C.uintptr_t(callback), C.uintptr_t(arg))
}
