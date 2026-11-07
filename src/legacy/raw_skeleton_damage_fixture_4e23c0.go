//go:build darwin && arm64

package legacy

/*
#include "GAME3_3.h"
#include <stddef.h>
#include <stdint.h>
typedef int32_t (*raw_skeleton_native_fn_4e23c0)(nox_object_t*,nox_object_t*,nox_object_t*,int32_t,int32_t);
_Static_assert(_Generic(&sub_4E23C0,raw_skeleton_native_fn_4e23c0:1,default:0),"Skeleton full native pointer ABI");
static int32_t raw_skeleton_fixture_call_4e23c0(void*t,void*s,void*w,int32_t d,int32_t k) {
 raw_skeleton_native_fn_4e23c0 fn=&sub_4E23C0;
 return fn((nox_object_t*)t,(nox_object_t*)s,(nox_object_t*)w,d,k);
}
static uintptr_t raw_skeleton_fixture_offset_4e23c0(int which) {
 switch(which) {
 case 0:return offsetof(nox_object_t,x);
 case 1:return offsetof(nox_object_t,prev_x);
 case 2:return offsetof(nox_object_t,prev_y);
 case 3:return offsetof(nox_object_t,direction1);
 case 4:return offsetof(nox_object_t,data_update);
 default:return UINTPTR_MAX;
 }
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func rawSkeletonFixtureCall4E23C0(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	result := C.raw_skeleton_fixture_call_4e23c0(target, source, weapon, C.int32_t(damage), C.int32_t(typ))
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	return int32(result)
}

func rawSkeletonFixtureCObjectOffsets4E23C0() (out [5]uintptr) {
	for i := range out {
		out[i] = uintptr(C.raw_skeleton_fixture_offset_4e23c0(C.int(i)))
	}
	return out
}
