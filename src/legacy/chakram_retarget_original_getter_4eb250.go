package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "chakram_update_53dcc0.h"
#include "memmap.h"
extern uint32_t dword_5d4594_1567928;
static void chakram_retarget_original_seed_4eb250(nox_object_t* source) {
 dword_5d4594_1567928=0;
 *getMemU32Ptr(0x5D4594,1567924)=(uint32_t)(uintptr_t)source;
 *getMemU32Ptr(0x5D4594,1567840)=0;
 *getMemU32Ptr(0x5D4594,1567932)=0;
 *getMemU32Ptr(0x5D4594,1567836)=0x4b189680;
}
static uint32_t chakram_retarget_original_selected_4eb250(void) { return dword_5d4594_1567928; }
static float2* chakram_retarget_original_bad_point_4eb250(int kind,int offset) {
 return (float2*)((uintptr_t)(kind==-1 ? 0 : 0x41000000)+(unsigned)offset);
}
static nox_object_t* chakram_retarget_old_4eb250(int (*fn)(int), nox_object_t* source) {
 return (nox_object_t*)(uintptr_t)(uint32_t)fn((int)(uintptr_t)source);
}
static nox_object_t* chakram_retarget_native_4eb250(nox_object_t* (*fn)(nox_object_t*), nox_object_t* source) { return fn(source); }
static void chakram_candidate_old_4eb250(void (*fn)(float*,int), nox_object_t* candidate, float2* point) { fn((float*)candidate,(int)(uintptr_t)point); }
static void chakram_candidate_native_4eb250(void (*fn)(nox_object_t*,float2*), nox_object_t* candidate, float2* point) { fn(candidate,point); }
static nox_object_t* chakram_retarget_original_public_4eb250(int entry, nox_object_t* object, float2* point) {
 if (entry==0) return _Generic(&sub_4EB250,int (*)(int):chakram_retarget_old_4eb250,nox_object_t* (*)(nox_object_t*):chakram_retarget_native_4eb250)(&sub_4EB250,object);
 _Generic(&sub_4EB340,void (*)(float*,int):chakram_candidate_old_4eb250,void (*)(nox_object_t*,float2*):chakram_candidate_native_4eb250)(&sub_4EB340,object,point);
 return NULL;
}
static int chakram_retarget_original_abi_4eb250(int entry) {
 if (entry==0) return _Generic(&sub_4EB250,int (*)(int):32,nox_object_t* (*)(nox_object_t*):64);
 return _Generic(&sub_4EB340,void (*)(float*,int):32,void (*)(nox_object_t*,float2*):64);
}
static uintptr_t chakram_retarget_original_layout_4eb250(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,obj_flags);
 case 2:return offsetof(nox_object_t,x);
 case 3:return offsetof(nox_object_t,vel_x);
 case 4:return offsetof(nox_object_t,owner);
 case 5:return offsetof(nox_object_t,speed_cur);
 case 6:return offsetof(nox_object_t,data_update);
 case 7:return sizeof(nox_object_t);
 case 8:return offsetof(nox_chakram_update_data_t,last_hit);
 case 9:return offsetof(nox_chakram_update_data_t,return_state);
 default:return sizeof(nox_chakram_update_data_t);
 }
}
static void chakram_retarget_original_controls_4eb250(nox_object_t* source, uintptr_t* out) {
 nox_chakram_update_data_t* update=(nox_chakram_update_data_t*)source->data_update;
 out[0]=(uintptr_t)source->owner;out[1]=(uintptr_t)update;out[2]=(uintptr_t)update->last_hit;
}
static int chakram_retarget_original_map_4eb250(nox_object_t* candidate,nox_object_t* source) { return nox_xxx_mapCheck_537110(candidate,source); }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func chakramRetargetOriginalC4EB250(entry int, object *server.Object, point *types.Pointf) *server.Object {
	return (*server.Object)(unsafe.Pointer(C.chakram_retarget_original_public_4eb250(C.int(entry), (*C.nox_object_t)(unsafe.Pointer(object)), (*C.float2)(unsafe.Pointer(point)))))
}
func chakramRetargetOriginalABI4EB250(entry int) int {
	return int(C.chakram_retarget_original_abi_4eb250(C.int(entry)))
}
func chakramRetargetOriginalLayout4EB250() (out [11]uintptr) {
	for i := range out {
		out[i] = uintptr(C.chakram_retarget_original_layout_4eb250(C.int(i)))
	}
	return
}
func chakramRetargetOriginalControls4EB250(source *server.Object) (out [3]uintptr) {
	var raw [3]C.uintptr_t
	C.chakram_retarget_original_controls_4eb250((*C.nox_object_t)(unsafe.Pointer(source)), &raw[0])
	for i := range out {
		out[i] = uintptr(raw[i])
	}
	return
}
func chakramRetargetOriginalMap4EB250(candidate, source *server.Object) bool {
	return C.chakram_retarget_original_map_4eb250((*C.nox_object_t)(unsafe.Pointer(candidate)), (*C.nox_object_t)(unsafe.Pointer(source))) != 0
}
func chakramRetargetOriginalSeed4EB250(source *server.Object) {
	C.chakram_retarget_original_seed_4eb250((*C.nox_object_t)(unsafe.Pointer(source)))
}
func chakramRetargetOriginalSelected4EB250() uint32 {
	return uint32(C.chakram_retarget_original_selected_4eb250())
}
func chakramRetargetOriginalBadPoint4EB250(kind, offset int) *types.Pointf {
	return (*types.Pointf)(unsafe.Pointer(C.chakram_retarget_original_bad_point_4eb250(C.int(kind), C.int(offset))))
}

var chakramRetargetNativeScratch4EB250 server.ChakramRetargetState4EB250
var chakramRetargetNativeRuntimeOverride4EB250 *server.ChakramRetargetRuntime4EB250
