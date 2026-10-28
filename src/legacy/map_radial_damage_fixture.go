package legacy

/*
#include "defs.h"
#include "GAME3_3.h"
#include <string.h>
#if defined(__i386__) && defined(__GNUC__)
#define NOX_RADIAL_LABEL_INNER(x) #x
#define NOX_RADIAL_LABEL(x) NOX_RADIAL_LABEL_INNER(x)
// Integer arguments have the same DWORD stack representation as float on
// i386. Let the C compiler own the complete call frame, then tail-jump to
// the real float entry without an x87 load or an extra return frame.
extern void nox_radial_damage_raw_probe(float2*,uint32_t,uint32_t,int,int,nox_object_t*,nox_object_t*);
__asm__(".text\n.globl " NOX_RADIAL_LABEL(__USER_LABEL_PREFIX__) "nox_radial_damage_raw_probe\n"
 NOX_RADIAL_LABEL(__USER_LABEL_PREFIX__) "nox_radial_damage_raw_probe:\n"
 "jmp " NOX_RADIAL_LABEL(__USER_LABEL_PREFIX__) "nox_xxx_mapDamageUnitsAround_4E25B0\n");
#undef NOX_RADIAL_LABEL
#undef NOX_RADIAL_LABEL_INNER
#endif
static void nox_radial_damage_probe(float2* center,uint32_t outer_bits,uint32_t inner_bits,int damage,int typ,nox_object_t* source,nox_object_t* excluded) {
#if defined(__i386__) && defined(__GNUC__)
 nox_radial_damage_raw_probe(center,outer_bits,inner_bits,damage,typ,source,excluded);
#else
 void (*call)(float*,float,float,int,int,nox_object_t*,nox_object_t*)=nox_xxx_mapDamageUnitsAround_4E25B0;
 float outer,inner;memcpy(&outer,&outer_bits,4);memcpy(&inner,&inner_bits,4);
 call((float*)center,outer,inner,damage,typ,source,excluded);
#endif
}
static int nox_radial_capture_count, nox_radial_capture_damage, nox_radial_capture_type, nox_radial_capture_shift;
static nox_object_t *nox_radial_capture_target, *nox_radial_capture_source, *nox_radial_capture_weapon;
static int nox_radial_capture(nox_object_t* target,nox_object_t* source,nox_object_t* weapon,int damage,int typ) {
 ++nox_radial_capture_count;nox_radial_capture_damage=damage;nox_radial_capture_type=typ;
 nox_radial_capture_target=target;nox_radial_capture_source=source;nox_radial_capture_weapon=weapon;
 if(nox_radial_capture_shift && nox_radial_capture_count==1 && source){source->x=500;source->y=500;}
 return 1;
}
static void* nox_radial_capture_pointer(int shift) {
 nox_radial_capture_count=0;nox_radial_capture_damage=0;nox_radial_capture_type=0;nox_radial_capture_shift=shift;
 nox_radial_capture_target=0;nox_radial_capture_source=0;nox_radial_capture_weapon=0;
 return nox_radial_capture;
}
static int nox_radial_capture_word(int n) { return n==0?nox_radial_capture_count:n==1?nox_radial_capture_damage:nox_radial_capture_type; }
static nox_object_t* nox_radial_capture_object(int n) {return n==0?nox_radial_capture_target:n==1?nox_radial_capture_source:nox_radial_capture_weapon;}
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"math"
	"unsafe"
)

func MapRadialDamageCProbe4E25B0(center *types.Pointf, outer, inner float32, damage int32, typ object.DamageType, source, excluded *server.Object) {
	C.nox_radial_damage_probe((*C.float2)(unsafe.Pointer(center)), C.uint32_t(math.Float32bits(outer)), C.uint32_t(math.Float32bits(inner)), C.int(damage), C.int(typ), asObjectC(source), asObjectC(excluded))
}

func RadialDamageCapturePointer4E25B0(shift bool) unsafe.Pointer {
	var n C.int
	if shift {
		n = 1
	}
	return C.nox_radial_capture_pointer(n)
}
func RadialDamageCaptureResult4E25B0() (count, damage, typ int32, target, source, weapon *server.Object) {
	return int32(C.nox_radial_capture_word(0)), int32(C.nox_radial_capture_word(1)), int32(C.nox_radial_capture_word(2)), asObjectS(C.nox_radial_capture_object(0)), asObjectS(C.nox_radial_capture_object(1)), asObjectS(C.nox_radial_capture_object(2))
}

// RadialDamageCaptureCall4E25B0 is the typed native bridge registered by the
// integration fixture. The production dispatcher rejects unknown PE32 calls.
func RadialDamageCaptureCall4E25B0(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
	return C.nox_radial_capture(asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ)) != 0
}
