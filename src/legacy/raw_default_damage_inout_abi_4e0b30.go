//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <string.h>
#include "defs.h"
typedef struct {uintptr_t words[5];uint32_t argc,response,mutation;} rdd_inout_abi_4e0b30;
static _Thread_local rdd_inout_abi_4e0b30* rdd_inout_current_4e0b30;
static void rdd_inout_write_4e0b30(int32_t*damage){memcpy(damage,&rdd_inout_current_4e0b30->mutation,4);}
typedef int32_t(*rdd_inout3_fn_4e0b30)(nox_object_t*,nox_object_t*,int32_t*);
typedef void(*rdd_inout4_fn_4e0b30)(nox_object_t*,int32_t*,uint32_t,nox_object_t*);
typedef int32_t(*rdd_inout5_fn_4e0b30)(nox_object_t*,nox_object_t*,nox_object_t*,int32_t*,uint32_t);
static int32_t rdd_inout3_4e0b30(nox_object_t*a,nox_object_t*b,int32_t*d){rdd_inout_abi_4e0b30*c=rdd_inout_current_4e0b30;c->argc=3;c->words[0]=(uintptr_t)a;c->words[1]=(uintptr_t)b;c->words[2]=(uintptr_t)d;rdd_inout_write_4e0b30(d);return (int32_t)c->response;}
static void rdd_inout4_4e0b30(nox_object_t*a,int32_t*d,uint32_t k,nox_object_t*b){rdd_inout_abi_4e0b30*c=rdd_inout_current_4e0b30;c->argc=4;c->words[0]=(uintptr_t)a;c->words[1]=(uintptr_t)d;c->words[2]=k;c->words[3]=(uintptr_t)b;rdd_inout_write_4e0b30(d);}
static int32_t rdd_inout5_4e0b30(nox_object_t*a,nox_object_t*b,nox_object_t*w,int32_t*d,uint32_t k){rdd_inout_abi_4e0b30*c=rdd_inout_current_4e0b30;c->argc=5;c->words[0]=(uintptr_t)a;c->words[1]=(uintptr_t)b;c->words[2]=(uintptr_t)w;c->words[3]=(uintptr_t)d;c->words[4]=k;rdd_inout_write_4e0b30(d);return (int32_t)c->response;}
// The C4 helper is void; wrapper response is an independent packet, not its C return.
static uint32_t rdd_inout_transport_4e0b30(uint32_t argc,void*a,void*b,void*w,int32_t*d,uint32_t k,uint32_t response,uint32_t mutation,uintptr_t*out){
 rdd_inout_abi_4e0b30 frame={0};rdd_inout_abi_4e0b30*parent=rdd_inout_current_4e0b30;frame.response=response;frame.mutation=mutation;rdd_inout_current_4e0b30=&frame;
 uint32_t result=0;if(argc==3){rdd_inout3_fn_4e0b30 fn=rdd_inout3_4e0b30;result=fn(a,b,d);}else if(argc==4){rdd_inout4_fn_4e0b30 fn=rdd_inout4_4e0b30;fn(a,d,k,b);result=frame.response;}else if(argc==5){rdd_inout5_fn_4e0b30 fn=rdd_inout5_4e0b30;result=fn(a,b,w,d,k);}
 memcpy(out,frame.words,sizeof(frame.words));out[5]=frame.argc;rdd_inout_current_4e0b30=parent;return result;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

func rawDefaultDamageInoutABI4E0B30(site uint32) (argc, damageIndex int, err error) {
	switch site {
	case 0x4e0f77, 0x4e0fef:
		return 5, 3, nil
	case 0x4e117c:
		return 3, 2, nil
	case 0x4e11f2:
		return 4, 1, nil
	default:
		return 0, 0, fmt.Errorf("undeclared native damage inout site %#x", site)
	}
}

// Objects are already in the original service's pointer-argument order. This
// function never chooses weapon/source based on parent gameplay conditions.
func rawDefaultDamageTransportInoutABI4E0B30(site uint32, objects [3]unsafe.Pointer, damage *int32, kind, response, mutation uint32) (words [5]uintptr, argc int, result uint32, err error) {
	count, _, e := rawDefaultDamageInoutABI4E0B30(site)
	if e != nil {
		return words, 0, 0, e
	}
	var out [6]uintptr
	got := C.rdd_inout_transport_4e0b30(C.uint32_t(count), objects[0], objects[1], objects[2], (*C.int32_t)(unsafe.Pointer(damage)), C.uint32_t(kind), C.uint32_t(response), C.uint32_t(mutation), (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	copy(words[:], out[:5])
	runtime.KeepAlive(objects)
	runtime.KeepAlive(damage)
	return words, int(out[5]), uint32(got), nil
}
