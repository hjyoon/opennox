package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME1.h"
#include "GAME4_3.h"
#include "GAME5.h"
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } edible_item_544a00;
typedef struct {
 uint32_t prefix[75]; void* waypoints[16]; uint32_t fields91[12];
 uint16_t health103[32]; void* animation119; uint32_t flags120;
 void* monster121; void* sound122; uint32_t fields123[13];
 int8_t index136; uint8_t pad136; uint16_t pad136b; uint32_t frame137;
 edible_item_544a00 stack138[24];
} edible_update_544a00;
static void* edible_root_old_544a00(int* (*fn)(int), nox_object_t* unit) { return fn((int)(uintptr_t)unit); }
static void* edible_root_native_544a00(void* (*fn)(nox_object_t*), nox_object_t* unit) { return fn(unit); }
static nox_object_t* edible_search_old_544a00(int (*fn)(nox_object_t*,float), nox_object_t* unit, float radius) { return (nox_object_t*)(uintptr_t)(uint32_t)fn(unit,radius); }
static nox_object_t* edible_search_native_544a00(nox_object_t* (*fn)(nox_object_t*,float), nox_object_t* unit, float radius) { return fn(unit,radius); }
static void edible_callback_old_544a00(void (*fn)(int,int), nox_object_t* candidate, nox_object_t* unit) { fn((int)(uintptr_t)candidate,(int)(uintptr_t)unit); }
static void edible_callback_native_544a00(void (*fn)(nox_object_t*,nox_object_t*), nox_object_t* candidate, nox_object_t* unit) { fn(candidate,unit); }
static void* edible_original_public_544a00(int entry, nox_object_t* unit, nox_object_t* candidate, float radius) {
 if (entry==0) return _Generic(&nox_xxx_mobRetreatCheckEdibles_5455E0,int* (*)(int):edible_root_old_544a00,void* (*)(nox_object_t*):edible_root_native_544a00)(&nox_xxx_mobRetreatCheckEdibles_5455E0,unit);
 if (entry==1) return _Generic(&nox_xxx_mobSearchEdible_544A00,int (*)(nox_object_t*,float):edible_search_old_544a00,nox_object_t* (*)(nox_object_t*,float):edible_search_native_544a00)(&nox_xxx_mobSearchEdible_544A00,unit,radius);
 _Generic(&nox_xxx_mobSearchEdible2_544A40,void (*)(int,int):edible_callback_old_544a00,void (*)(nox_object_t*,nox_object_t*):edible_callback_native_544a00)(&nox_xxx_mobSearchEdible2_544A40,candidate,unit);
 return NULL;
}
static int edible_original_abi_544a00(int entry) {
 if (entry==0) return _Generic(&nox_xxx_mobRetreatCheckEdibles_5455E0,int* (*)(int):32,void* (*)(nox_object_t*):64);
 if (entry==1) return _Generic(&nox_xxx_mobSearchEdible_544A00,int (*)(nox_object_t*,float):32,nox_object_t* (*)(nox_object_t*,float):64);
 return _Generic(&nox_xxx_mobSearchEdible2_544A40,void (*)(int,int):32,void (*)(nox_object_t*,nox_object_t*):64);
}
static int edible_original_flags_544a00(uint32_t mask) { return nox_common_gameFlags_check_40A5C0((int)mask); }
static uintptr_t edible_original_layout_544a00(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,obj_subclass);
 case 2:return offsetof(nox_object_t,x);
 case 3:return offsetof(nox_object_t,poison_540);
 case 4:return offsetof(nox_object_t,data_update);
 case 5:return sizeof(nox_object_t);
 case 6:return offsetof(edible_update_544a00,index136);
 case 7:return offsetof(edible_update_544a00,stack138);
 case 8:return sizeof(edible_item_544a00);
 case 9:return offsetof(edible_item_544a00,args);
 default:return offsetof(edible_item_544a00,started);
 }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func edibleOriginalC544A00(entry int, unit, candidate *server.Object, radius float32) unsafe.Pointer {
	return C.edible_original_public_544a00(C.int(entry), (*C.nox_object_t)(unsafe.Pointer(unit)), (*C.nox_object_t)(unsafe.Pointer(candidate)), C.float(radius))
}
func edibleOriginalABI544A00(entry int) int { return int(C.edible_original_abi_544a00(C.int(entry))) }
func edibleOriginalLayout544A00() (out [11]uintptr) {
	for i := range out {
		out[i] = uintptr(C.edible_original_layout_544a00(C.int(i)))
	}
	return
}

var edibleNativeScratch544A00 server.MonsterEdibleSearchState544A00
var edibleNativeRuntimeOverride544A00 *server.MonsterEdibleNativeRuntime544A00

func edibleOriginalNativeFlags544A00(mask uint32) bool {
	return C.edible_original_flags_544a00(C.uint32_t(mask)) != 0
}
