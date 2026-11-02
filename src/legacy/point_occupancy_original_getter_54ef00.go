package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
static int occupancy_parent_legacy_54ef00(int (*fn)(float*),float2* point) { return fn((float*)point); }
static int occupancy_parent_native_54ef00(int (*fn)(float2*),float2* point) { return fn(point); }
static int occupancy_parent_abi_54ef00(void) { return _Generic(&sub_54EF00,int (*)(float*):32,int (*)(float2*):64); }
static int occupancy_parent_public_54ef00(float2* point) {
 return _Generic(&sub_54EF00,int (*)(float*):occupancy_parent_legacy_54ef00,int (*)(float2*):occupancy_parent_native_54ef00)(&sub_54EF00,point);
}
static void occupancy_candidate_legacy_54ef60(void (*fn)(float*,int),nox_object_t* unit,float2* point) { fn((float*)unit,(int)(uintptr_t)point); }
static void occupancy_candidate_native_54ef60(void (*fn)(nox_object_t*,float2*),nox_object_t* unit,float2* point) { fn(unit,point); }
static int occupancy_candidate_abi_54ef60(void) { return _Generic(&sub_54EF60,void (*)(float*,int):32,void (*)(nox_object_t*,float2*):64); }
static void occupancy_candidate_public_54ef60(nox_object_t* unit,float2* point) {
 _Generic(&sub_54EF60,void (*)(float*,int):occupancy_candidate_legacy_54ef60,void (*)(nox_object_t*,float2*):occupancy_candidate_native_54ef60)(&sub_54EF60,unit,point);
}
static uintptr_t occupancy_layout_54ef60(int field) {
 return field==0 ? offsetof(nox_object_t,obj_class) : offsetof(nox_object_t,field_5);
}
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

type pointOccupancyOriginalRuntime54EF00 struct {
	Each  func(types.Rectf, func(*server.Object) bool)
	Load  func() uint32
	Store func(uint32)
}

var pointOccupancyRuntimeOverride54EF00 *pointOccupancyOriginalRuntime54EF00

func occupancyOriginalParentABI54EF00() int    { return int(C.occupancy_parent_abi_54ef00()) }
func occupancyOriginalCandidateABI54EF60() int { return int(C.occupancy_candidate_abi_54ef60()) }
func occupancyOriginalParentC54EF00(point *types.Pointf) uint32 {
	return uint32(C.occupancy_parent_public_54ef00((*C.float2)(unsafe.Pointer(point))))
}
func occupancyOriginalCandidateC54EF60(unit *server.Object, point *types.Pointf) {
	C.occupancy_candidate_public_54ef60((*C.nox_object_t)(unsafe.Pointer(unit)), (*C.float2)(unsafe.Pointer(point)))
}
func occupancyOriginalLayout54EF60() [2]uintptr {
	return [2]uintptr{uintptr(C.occupancy_layout_54ef60(0)), uintptr(C.occupancy_layout_54ef60(1))}
}
