package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
static int point_shape_legacy_547db0(int (*fn)(int,float2*),nox_object_t* unit,float2* point) { return fn((int)(uintptr_t)unit,point); }
static int point_shape_native_547db0(int (*fn)(nox_object_t*,float2*),nox_object_t* unit,float2* point) { return fn(unit,point); }
static int point_shape_abi_547db0(void) { return _Generic(&sub_547DB0,int (*)(int,float2*):32,int (*)(nox_object_t*,float2*):64); }
static int point_shape_public_547db0(nox_object_t* unit,float2* point) {
 return _Generic(&sub_547DB0,int (*)(int,float2*):point_shape_legacy_547db0,int (*)(nox_object_t*,float2*):point_shape_native_547db0)(&sub_547DB0,unit,point);
}
static uintptr_t point_shape_layout_547db0(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,new_x);
 case 1:return offsetof(nox_object_t,new_y);
 case 2:return offsetof(nox_object_t,shape);
 case 3:return offsetof(nox_object_t,shape)+offsetof(nox_shape,circle_r2);
 default:return sizeof(float2);
 }
}
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func pointShapeOriginalC547DB0(unit *server.Object, point *types.Pointf) int {
	return int(C.point_shape_public_547db0((*C.nox_object_t)(unsafe.Pointer(unit)), (*C.float2)(unsafe.Pointer(point))))
}
func pointShapeOriginalABI547DB0() int { return int(C.point_shape_abi_547db0()) }
func pointShapeOriginalLayout547DB0() (out [5]uintptr) {
	for i := range out {
		out[i] = uintptr(C.point_shape_layout_547db0(C.int(i)))
	}
	return out
}
