package legacy

/*
#include "defs.h"
#include "GAME3_3.h"
#include <string.h>
static void nox_radial_damage_probe(float2* center,uint32_t outer_bits,uint32_t inner_bits,int damage,int typ,nox_object_t* source,nox_object_t* excluded) {
 void (*call)(float*,float,float,int,int,nox_object_t*,nox_object_t*)=nox_xxx_mapDamageUnitsAround_4E25B0;
#if defined(__i386__) && defined(__GNUC__)
 // Reproduce GAME.EXE's raw DWORD pushes. MinGW's float copies, including
 // the cgo probe's own argument marshalling, otherwise quiet the input
 // before the production C entry can see it.
 struct {float2* center;uint32_t outer,inner;int damage,typ;nox_object_t *source,*excluded;void* call;uintptr_t stack;} frame =
  {center,outer_bits,inner_bits,damage,typ,source,excluded,(void*)call,0};
 __asm__ volatile("movl %%esp,32(%0)\n\tandl $-16,%%esp\n\tsubl $4,%%esp\n\tpushl 24(%0)\n\tpushl 20(%0)\n\tpushl 16(%0)\n\tpushl 12(%0)\n\tpushl 8(%0)\n\tpushl 4(%0)\n\tpushl 0(%0)\n\tcall *28(%0)\n\tmovl 32(%0),%%esp"
  : : "r"(&frame) : "eax","ecx","edx","memory","cc");
#else
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
