package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include <string.h>
extern uint32_t dword_5d4594_2488604;
extern uint32_t dword_5d4594_2488608;
#if defined(__i386__) && defined(__GNUC__)
#define NOX_VELOCITY_LABEL_INNER(x) #x
#define NOX_VELOCITY_LABEL(x) NOX_VELOCITY_LABEL_INNER(x)
// The compiler owns the call frame. The tail jump preserves the DWORD
// float representation without introducing an x87 load in the test caller.
extern int nox_velocity_raw_probe_5118a0(uint32_t);
__asm__(".text\n.globl " NOX_VELOCITY_LABEL(__USER_LABEL_PREFIX__) "nox_velocity_raw_probe_5118a0\n"
 NOX_VELOCITY_LABEL(__USER_LABEL_PREFIX__) "nox_velocity_raw_probe_5118a0:\n"
 "jmp " NOX_VELOCITY_LABEL(__USER_LABEL_PREFIX__) "nox_xxx_updateObjectsVelocity_5118A0\n");
#undef NOX_VELOCITY_LABEL
#undef NOX_VELOCITY_LABEL_INNER
#endif
static int nox_velocity_probe_5118a0(uint32_t word) {
#if defined(__i386__) && defined(__GNUC__)
 return nox_velocity_raw_probe_5118a0(word);
#else
 float step;memcpy(&step,&word,sizeof(step));
 int (*call)(float)=nox_xxx_updateObjectsVelocity_5118A0;
 return call(step);
#endif
}
static void nox_velocity_fixture_update_5118a0(nox_object_t* o) {(void)o;}
static void* nox_velocity_fixture_update_pointer_5118a0(void) {return nox_velocity_fixture_update_5118a0;}

// Check FLD m32fp directly on x86. Upstream Unicorn 2.1.4 incorrectly leaves
// a signaling NaN signaling after this load, affecting mixed-NaN oracles.
// Save and restore the complete x87 environment around the probe.
static uint64_t nox_velocity_load_nan_probe_5118a0(uint32_t word) {
#if (defined(__i386__) || defined(__x86_64__)) && defined(__GNUC__)
 unsigned char saved[28],raw[10];uint16_t cw=0xe3f,sw;uint64_t sig;uint16_t exp;
 float input;memcpy(&input,&word,sizeof(input));
 __asm__ volatile("fnstenv %0; fnclex; fldcw %3; flds %4; fnstsw %1; fstpt %2; fldenv %0"
  : "=m"(saved),"=m"(sw),"=m"(raw) : "m"(cw),"m"(input) : "st");
 memcpy(&sig,raw,8);memcpy(&exp,raw+8,2);
 uint32_t result=((uint32_t)(exp&0x8000)<<16)|0x7f800000u|(uint32_t)((sig>>40)&0x7fffff);
 return ((uint64_t)sw<<32)|result;
#else
 return 0;
#endif
}
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func ObjectsVelocityCProbe5118A0(step float32) int32 {
	return int32(C.nox_velocity_probe_5118a0(C.uint32_t(math.Float32bits(step))))
}

// ObjectsVelocityQueueFixture5118A0 isolates the real pending queue while
// retaining its production enqueue gates and all production velocity services.
func ObjectsVelocityQueueFixture5118A0(objects ...*server.Object) func() {
	old := pendingCollisionQueue537580
	oldHead, oldTail := C.dword_5d4594_2488604, C.dword_5d4594_2488608
	pendingCollisionQueue537580 = server.PendingCollisionQueue537580{}
	for _, o := range objects {
		pendingCollisionEnqueue537610(o)
	}
	return func() {
		pendingCollisionQueue537580 = old
		C.dword_5d4594_2488604, C.dword_5d4594_2488608 = oldHead, oldTail
	}
}

func ObjectsVelocityUpdatePointer5118A0() unsafe.Pointer {
	return C.nox_velocity_fixture_update_pointer_5118a0()
}

func velocityLoadNaNProbe5118A0(word uint32) uint64 {
	return uint64(C.nox_velocity_load_nan_probe_5118a0(C.uint32_t(word)))
}
