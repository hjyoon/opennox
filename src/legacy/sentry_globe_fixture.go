package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include <string.h>
static int nox_sentry_x87_available(void) {
#if defined(__i386__) || defined(__x86_64__)
 return 1;
#else
 return 0;
#endif
}
static uint32_t nox_sentry_x87_endpoint(uint32_t angle_bits,uint32_t origin_bits,int cosine) {
#if defined(__i386__) || defined(__x86_64__)
 float angle,origin,scale=600.0f,out;unsigned short old,cw=0xe3f;
 memcpy(&angle,&angle_bits,4);memcpy(&origin,&origin_bits,4);
 __asm__ volatile("fnstcw %0":"=m"(old));__asm__ volatile("fldcw %0"::"m"(cw));
 if(cosine)__asm__ volatile("flds %1; fcos; fmuls %2; fadds %3; fstps %0":"=m"(out):"m"(angle),"m"(scale),"m"(origin):"st","memory");
 else __asm__ volatile("flds %1; fsin; fmuls %2; fadds %3; fstps %0":"=m"(out):"m"(angle),"m"(scale),"m"(origin):"st","memory");
 __asm__ volatile("fldcw %0"::"m"(old));uint32_t bits;memcpy(&bits,&out,4);return bits;
#else
 return 0;
#endif
}
static void* nox_sentry_probe(int kind,nox_object_t* source,void* context) {
 switch(kind) {
 case 0: { void (*call)(nox_object_t*)=nox_xxx_updateSentryGlobe_510E60;call(source);return 0; }
 case 1: { nox_object_t* (*call)(nox_object_t*)=nox_xxx_sentryUpdateList_510FD0;return call(source); }
 case 2: { void (*call)(nox_object_t*,nox_sentry_ray_context*)=nox_xxx_sentry_511020;call(source,context);return 0; }
 case 3: sub_511250(2,source);return 0;
 case 4: sub_510E50();return 0;
 default: sub_511100(2);return 0;
 }
}
static int nox_sentry_packet_probe(int index,nox_object_t* source) { int (*call)(int,nox_object_t*)=sub_511250;return call(index,source); }
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func sentryGlobeProbe(kind int, source *server.Object, context unsafe.Pointer) *server.Object {
	return (*server.Object)(C.nox_sentry_probe(C.int(kind), asObjectC(source), context))
}

func sentryGlobePacketProbe(index int32, source *server.Object) int32 {
	return int32(C.nox_sentry_packet_probe(C.int(index), asObjectC(source)))
}

func sentryGlobeX87EndpointProbe(angle, origin uint32, cosine bool) (uint32, bool) {
	if C.nox_sentry_x87_available() == 0 {
		return 0, false
	}
	var c C.int
	if cosine {
		c = 1
	}
	return uint32(C.nox_sentry_x87_endpoint(C.uint32_t(angle), C.uint32_t(origin), c)), true
}

func sentryGlobeX87Available() bool { return C.nox_sentry_x87_available() != 0 }
