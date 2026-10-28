package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var sentryGlobeRawRuntime510E60 = func() server.SentryGlobeRawRuntime510E60 {
	return GetServer().S().SentryGlobeRuntime510E60()
}

//export nox_sentry_cluster_native
func nox_sentry_cluster_native(kind C.int, source *nox_object_t, context *C.nox_sentry_ray_context, index C.int) *nox_object_t {
	h := sentryGlobeRawRuntime510E60()
	switch kind {
	case 0:
		server.SentryGlobeUpdateRaw510E60(asObjectS(source), h)
	case 1:
		return asObjectC(server.SentryGlobeUnlinkRaw510FD0(asObjectS(source), h.List))
	case 2:
		server.SentryGlobeHitRaw511020(asObjectS(source), (*server.SentryGlobeRayContext511020)(unsafe.Pointer(context)), h)
	case 3:
		server.SentryGlobeResetRaw510E50(h.List)
	case 4:
		server.SentryGlobeSendRaw511100(int32(index), h)
	default:
		panic("unknown Sentry entry")
	}
	return nil
}

//export nox_sentry_packet_native
func nox_sentry_packet_native(index C.int, source *nox_object_t) C.int {
	return C.int(server.SentryGlobePacketRaw511250(int32(index), asObjectS(source), sentryGlobeRawRuntime510E60().Packet))
}
