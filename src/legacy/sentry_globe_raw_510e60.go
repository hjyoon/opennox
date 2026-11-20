package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var sentryGlobeRawRuntime510E60 = func() server.SentryGlobeRawRuntime510E60 {
	return GetServer().S().SentryGlobeRuntime510E60()
}

//export sub_510E50
func sub_510E50() {
	server.SentryGlobeResetRaw510E50(sentryGlobeRawRuntime510E60().List)
}

//export nox_xxx_updateSentryGlobe_510E60
func nox_xxx_updateSentryGlobe_510E60(source *C.nox_object_t) {
	server.SentryGlobeUpdateRaw510E60(asObjectS(source), sentryGlobeRawRuntime510E60())
}

//export nox_xxx_sentryUpdateList_510FD0
func nox_xxx_sentryUpdateList_510FD0(source *C.nox_object_t) *C.nox_object_t {
	return asObjectC(server.SentryGlobeUnlinkRaw510FD0(asObjectS(source), sentryGlobeRawRuntime510E60().List))
}

//export nox_xxx_sentry_511020
func nox_xxx_sentry_511020(candidate *C.nox_object_t, context *C.nox_sentry_ray_context) {
	server.SentryGlobeHitRaw511020(asObjectS(candidate), (*server.SentryGlobeRayContext511020)(unsafe.Pointer(context)), sentryGlobeRawRuntime510E60())
}

//export sub_511100
func sub_511100(index C.int) {
	server.SentryGlobeSendRaw511100(int32(index), sentryGlobeRawRuntime510E60())
}

//export nox_sentry_cluster_native
func nox_sentry_cluster_native(kind C.int, source *C.nox_object_t, context *C.nox_sentry_ray_context, index C.int) *C.nox_object_t {
	switch kind {
	case 0:
		nox_xxx_updateSentryGlobe_510E60(source)
	case 1:
		return nox_xxx_sentryUpdateList_510FD0(source)
	case 2:
		nox_xxx_sentry_511020(source, context)
	case 3:
		sub_510E50()
	case 4:
		sub_511100(index)
	default:
		panic("unknown Sentry entry")
	}
	return nil
}

//export sub_511250
func sub_511250(index C.int, source *C.nox_object_t) C.int {
	return C.int(server.SentryGlobePacketRaw511250(int32(index), asObjectS(source), sentryGlobeRawRuntime510E60().Packet))
}
