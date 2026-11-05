package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
typedef int32_t (*modifier_predicate_call_4e03d0)(void*,nox_object_t*,uintptr_t,nox_object_t*,uintptr_t,int32_t*);
typedef struct {
 uintptr_t fn,args[6];
 int32_t phase,before,after,result;
} modifier_predicate_trace_4e03d0;
static int modifier_predicate_tracing_4e03d0,modifier_predicate_trace_count_4e03d0;
static modifier_predicate_trace_4e03d0 modifier_predicate_trace_data_4e03d0[32];
static void modifier_predicate_trace_begin_4e03d0(void){modifier_predicate_trace_count_4e03d0=0;modifier_predicate_tracing_4e03d0=1;}
static void modifier_predicate_trace_end_4e03d0(void){modifier_predicate_tracing_4e03d0=0;}
static int modifier_predicate_trace_append_4e03d0(void* fn,int phase,void* first,const uintptr_t* a,void* output,int32_t before,int32_t after,int32_t result){
 if(!modifier_predicate_tracing_4e03d0)return -1;
 int index=modifier_predicate_trace_count_4e03d0++;
 if(index>=32)return -1;
 modifier_predicate_trace_4e03d0* r=&modifier_predicate_trace_data_4e03d0[index];
 r->fn=(uintptr_t)fn;r->phase=phase;r->before=before;r->after=after;r->result=result;
 r->args[0]=(uintptr_t)first;for(int i=0;i<4;i++)r->args[i+1]=a[i];r->args[5]=(uintptr_t)output;
 return index;
}
static int modifier_predicate_trace_count_get_4e03d0(void){return modifier_predicate_trace_count_4e03d0;}
static modifier_predicate_trace_4e03d0 modifier_predicate_trace_get_4e03d0(int index){return modifier_predicate_trace_data_4e03d0[index];}
static void* modifier_predicate_trace_root_4e03d0(int which){
 switch(which){case 0:return (void*)nox_xxx_inversionEffect_4E03D0;case 1:return (void*)nox_xxx_gripEffect_4E0480;default:return (void*)nox_xxx_unusedCheckGripEffect_4E03F0;}
}
// No game algorithm: call the supplied full-width callback and observe its
// actual ABI arguments and caller-owned scratch word when tracing is enabled.
static int32_t modifier_predicate_invoke_4e03d0(void* fn,void* effect,const uintptr_t* a,int32_t* out){
 int32_t before=modifier_predicate_tracing_4e03d0?*out:0;
 modifier_predicate_trace_append_4e03d0(fn,2,effect,a,out,before,0,0);
 int32_t result=((modifier_predicate_call_4e03d0)fn)(effect,(nox_object_t*)a[0],a[1],(nox_object_t*)a[2],a[3],out);
 modifier_predicate_trace_append_4e03d0(fn,3,effect,a,out,before,modifier_predicate_tracing_4e03d0?*out:0,result);
 return result;
}
static uintptr_t modifier_predicate_transport_seen_4e03d0[6];
static int32_t modifier_predicate_transport_stub_4e03d0(void* effect,nox_object_t* item,uintptr_t context,nox_object_t* target,uintptr_t other,int32_t* out){
 uintptr_t values[6]={(uintptr_t)effect,(uintptr_t)item,context,(uintptr_t)target,other,(uintptr_t)out};
 for(int i=0;i<6;i++)modifier_predicate_transport_seen_4e03d0[i]=values[i];
 *out=INT32_C(0x13579bdf);return INT32_C(0x2468ace0);
}
static void* modifier_predicate_transport_pointer_4e03d0(void){return (void*)modifier_predicate_transport_stub_4e03d0;}
static uintptr_t modifier_predicate_transport_get_4e03d0(int index){return modifier_predicate_transport_seen_4e03d0[index];}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// These field accesses and callback transports are frozen before any native
// predicate/traversal engine. The observer is absent in ordinary production.
type modifierPredicateValue4E03D0 struct {
	Word    uint32
	Pointer unsafe.Pointer
}

var modifierPredicateObserve4E03D0 func(write, complete bool, base unsafe.Pointer, kind, index int, value modifierPredicateValue4E03D0)

func modifierPredicateRead4E03D0(base unsafe.Pointer, kind, index int) (value modifierPredicateValue4E03D0) {
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(false, false, base, kind, index, value)
	}
	switch kind {
	case 0:
		value.Word = uint32((*server.ModifierEff)(base).DefendCollide88.Val)
	case 1:
		value.Word = uint32((*server.Object)(base).ObjClass)
	case 2:
		value.Pointer = (*server.Object)(base).InitData
	case 3:
		value.Pointer = unsafe.Pointer((*server.ModifierInitData)(base).Modifiers[index])
	case 4:
		value.Pointer = (*server.ModifierEff)(base).DefendCollide88.Fnc
	case 5:
		value.Word = *(*uint32)(base)
	default:
		panic("undeclared modifier predicate field")
	}
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(false, true, base, kind, index, value)
	}
	return
}

func modifierPredicateStore4E03D0(output *int32, word uint32) {
	value := modifierPredicateValue4E03D0{Word: word}
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(true, false, unsafe.Pointer(output), 5, 0, value)
	}
	*output = int32(word)
	if modifierPredicateObserve4E03D0 != nil {
		modifierPredicateObserve4E03D0(true, true, unsafe.Pointer(output), 5, 0, value)
	}
}

type modifierPredicateTrace4E03D0 struct {
	Function              uintptr
	Args                  [6]uintptr
	Phase                 int
	Before, After, Result int32
}

func modifierPredicateTracing4E03D0(enabled bool) {
	if enabled {
		C.modifier_predicate_trace_begin_4e03d0()
	} else {
		C.modifier_predicate_trace_end_4e03d0()
	}
}

func modifierPredicateTraceRecords4E03D0() (records []modifierPredicateTrace4E03D0) {
	count := int(C.modifier_predicate_trace_count_get_4e03d0())
	if count < 0 || count > 32 {
		panic("modifier predicate trace overflow")
	}
	for i := 0; i < count; i++ {
		r := C.modifier_predicate_trace_get_4e03d0(C.int(i))
		row := modifierPredicateTrace4E03D0{Function: uintptr(r.fn), Phase: int(r.phase), Before: int32(r.before), After: int32(r.after), Result: int32(r.result)}
		for j := range row.Args {
			row.Args[j] = uintptr(r.args[j])
		}
		records = append(records, row)
	}
	return
}

func modifierPredicateRootPointer4E03D0(which int) unsafe.Pointer {
	return C.modifier_predicate_trace_root_4e03d0(C.int(which))
}

func modifierPredicatesTracedPublic4E03D0(which int, argument unsafe.Pointer, opaque [4]uint32, output unsafe.Pointer) int32 {
	var args [4]C.uintptr_t
	for i, value := range opaque {
		args[i] = C.uintptr_t(value)
	}
	fn := modifierPredicateRootPointer4E03D0(which)
	C.modifier_predicate_trace_append_4e03d0(fn, 0, argument, &args[0], output, 0, 0, 0)
	result := modifierPredicatesPublic4E03D0(which, argument, opaque, output)
	C.modifier_predicate_trace_append_4e03d0(fn, 1, argument, &args[0], output, 0, 0, C.int32_t(result))
	return result
}

func modifierPredicateInvoke4E03D0(fn, effect unsafe.Pointer, opaque [4]uintptr, output *int32) int32 {
	var args [4]C.uintptr_t
	for i, value := range opaque {
		args[i] = C.uintptr_t(value)
	}
	return int32(C.modifier_predicate_invoke_4e03d0(fn, effect, &args[0], (*C.int32_t)(unsafe.Pointer(output))))
}

func modifierPredicateTransportControl4E03D0(effect unsafe.Pointer, opaque [4]uintptr, output *int32) (result int32, actual [6]uintptr) {
	result = modifierPredicateInvoke4E03D0(C.modifier_predicate_transport_pointer_4e03d0(), effect, opaque, output)
	for i := range actual {
		actual[i] = uintptr(C.modifier_predicate_transport_get_4e03d0(C.int(i)))
	}
	return
}
