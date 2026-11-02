package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } roam_lifecycle_item_545790;
typedef struct {
 uint32_t prefix[75]; void* waypoints[16]; uint32_t fields91[12];
 uint16_t health103[32]; void* animation119; uint32_t flags120;
 void* monster121; void* sound122; uint32_t fields123[13];
 int8_t index136; uint8_t pad136; uint16_t pad136b; uint32_t frame137;
 roam_lifecycle_item_545790 stack138[24];
} roam_lifecycle_update_545790;
static void* roam_start_old_545790(long long (*fn)(int), nox_object_t* unit) { return (void*)(uintptr_t)fn((int)(uintptr_t)unit); }
static void* roam_cancel_old_545790(int (*fn)(int), nox_object_t* unit) { return (void*)(uintptr_t)(uint32_t)fn((int)(uintptr_t)unit); }
static void* roam_lifecycle_native_545790(void* (*fn)(nox_object_t*), nox_object_t* unit) { return fn(unit); }
static void* roam_lifecycle_original_545790(int entry, nox_object_t* unit) {
 if (!entry) return _Generic(&sub_545790,long long (*)(int):roam_start_old_545790,void* (*)(nox_object_t*):roam_lifecycle_native_545790)(&sub_545790,unit);
 return _Generic(&sub_5457C0,int (*)(int):roam_cancel_old_545790,void* (*)(nox_object_t*):roam_lifecycle_native_545790)(&sub_5457C0,unit);
}
static int roam_lifecycle_abi_545790(int entry) {
 if (!entry) return _Generic(&sub_545790,long long (*)(int):32,void* (*)(nox_object_t*):64);
 return _Generic(&sub_5457C0,int (*)(int):32,void* (*)(nox_object_t*):64);
}
static uintptr_t roam_lifecycle_layout_545790(int field) {
 switch (field) {
 case 0: return offsetof(nox_object_t,data_update);
 case 1: return sizeof(nox_object_t);
 case 2: return offsetof(roam_lifecycle_update_545790,waypoints);
 case 3: return offsetof(roam_lifecycle_update_545790,fields91);
 case 4: return offsetof(roam_lifecycle_update_545790,index136);
 case 5: return offsetof(roam_lifecycle_update_545790,stack138);
 case 6: return sizeof(roam_lifecycle_item_545790);
 default: return offsetof(roam_lifecycle_item_545790,args);
 }
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func roamLifecycleOriginalC545790(entry int, unit *server.Object) unsafe.Pointer {
	return C.roam_lifecycle_original_545790(C.int(entry), (*C.nox_object_t)(unsafe.Pointer(unit)))
}
func roamLifecycleOriginalABI545790(entry int) int {
	return int(C.roam_lifecycle_abi_545790(C.int(entry)))
}
func roamLifecycleOriginalLayout545790() (out [8]uintptr) {
	for i := range out {
		out[i] = uintptr(C.roam_lifecycle_layout_545790(C.int(i)))
	}
	return
}
