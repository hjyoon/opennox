package legacy

/*
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*arachnaphobia_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*);
_Static_assert(_Generic(&nox_xxx_spellArachna_52DC80,arachnaphobia_callback_C:1,default:0),"native five-argument Arachnaphobia ABI");
static void arachnaphobia_layout_C(uintptr_t* v){
 v[0]=offsetof(nox_object_t,obj_class);v[1]=offsetof(nox_object_t,x);v[2]=offsetof(nox_object_t,y);v[3]=offsetof(nox_object_t,data_update);
 v[4]=offsetof(nox_spell_accept_arg_t,obj);v[5]=offsetof(nox_spell_accept_arg_t,x);v[6]=offsetof(nox_spell_accept_arg_t,y);
 v[7]=offsetof(nox_player_update_data_t,player);v[8]=offsetof(nox_playerInfo,playerInd);
}
static int arachnaphobia_call_C(int route,int id,nox_object_t* second,nox_object_t* owner,nox_object_t* caster,void* arg){
 if(route==1){arachnaphobia_callback_C f=nox_xxx_spellArachna_52DC80;return f(id,second,owner,caster,arg);}
 return nox_xxx_spellArachna_52DC80(id,second,owner,caster,arg);
}
typedef int(*trace_ray9_callback_C)(float4*);
_Static_assert(_Generic(&nox_xxx_traceRay_5374B0,trace_ray9_callback_C:1,default:0),"native trace wrapper ABI");
static void trace_ray9_layout_C(uintptr_t* v){
 v[0]=sizeof(float4);v[1]=offsetof(float4,field_0);v[2]=offsetof(float4,field_4);v[3]=offsetof(float4,field_8);v[4]=offsetof(float4,field_C);
}
static int trace_ray9_call_C(int route,float4* line){
 if(route==1){trace_ray9_callback_C f=nox_xxx_traceRay_5374B0;return f(line);}
 return nox_xxx_traceRay_5374B0(line);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func arachnaphobiaLayoutC() (out [9]uintptr) {
	C.arachnaphobia_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func arachnaphobiaActualCall(route int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Nox_xxx_spellArachna_52DC80(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.arachnaphobia_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg)))
}

func traceRay9LayoutC() (out [5]uintptr) {
	C.trace_ray9_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func traceRay9ActualCall(route int, line unsafe.Pointer) int32 {
	if route == 2 {
		return int32(nox_xxx_traceRay_5374B0((*C.float4)(line)))
	}
	return int32(C.trace_ray9_call_C(C.int(route), (*C.float4)(line)))
}

func traceRay9FixtureService(call func(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer, int8) int32) func() {
	previous := traceRayMapCall5374B0
	traceRayMapCall5374B0 = func(line *C.float4, position *C.float2, grid *C.int2, flags C.char) int32 {
		return call(unsafe.Pointer(line), unsafe.Pointer(position), unsafe.Pointer(grid), int8(flags))
	}
	return func() { traceRayMapCall5374B0 = previous }
}
