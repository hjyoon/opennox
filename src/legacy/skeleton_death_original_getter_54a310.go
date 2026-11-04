package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5.h"
static int skeleton_old_call_54a310(int (*fn)(int),nox_object_t* u) {return fn((int)(uintptr_t)u);}
static int skeleton_native_call_54a310(int (*fn)(nox_object_t*),nox_object_t* u) {return fn(u);}
#define SKELETON_CALL(fn,u) _Generic(&(fn),int(*)(int):skeleton_old_call_54a310,int(*)(nox_object_t*):skeleton_native_call_54a310)(&(fn),u)
static void skeleton_old_debris_54a310(void(*fn)(int),nox_object_t* u) {fn((int)(uintptr_t)u);}
static void skeleton_native_debris_54a310(void(*fn)(nox_object_t*),nox_object_t* u) {fn(u);}
static void skeleton_old_drop_54a310(void(*fn)(int,char*,const char*,const char*,const char*,const char*,int),nox_object_t* u,char* type,char** mods,int charge) {fn((int)(uintptr_t)u,type,mods[0],mods[1],mods[2],mods[3],charge);}
static void skeleton_native_drop_54a310(void(*fn)(nox_object_t*,char*,char*,char*,char*,char*,int),nox_object_t* u,char* type,char** mods,int charge) {fn(u,type,mods[0],mods[1],mods[2],mods[3],charge);}
static int skeleton_public_54a310(int entry,nox_object_t* u,char* type,char** mods,int charge) {
 switch(entry) {
 case 0:return SKELETON_CALL(sub_54A310,u);
 case 1:_Generic(&sub_54A390,void(*)(int,char*,const char*,const char*,const char*,const char*,int):skeleton_old_drop_54a310,void(*)(nox_object_t*,char*,char*,char*,char*,char*,int):skeleton_native_drop_54a310)(&sub_54A390,u,type,mods,charge);return 1;
 case 2:_Generic(&sub_54A4C0,void(*)(int):skeleton_old_debris_54a310,void(*)(nox_object_t*):skeleton_native_debris_54a310)(&sub_54A4C0,u);return 1;
 case 3:return SKELETON_CALL(sub_54A750,u);
 default:return -1;
 }
}
static int skeleton_abi_54a310(int entry) {
 switch(entry) {
 case 0:return _Generic(&sub_54A310,int(*)(int):32,int(*)(nox_object_t*):64);
 case 1:return _Generic(&sub_54A390,void(*)(int,char*,const char*,const char*,const char*,const char*,int):32,void(*)(nox_object_t*,char*,char*,char*,char*,char*,int):64);
 case 2:return _Generic(&sub_54A4C0,void(*)(int):32,void(*)(nox_object_t*):64);
 case 3:return _Generic(&sub_54A750,int(*)(int):32,int(*)(nox_object_t*):64);
 default:return 0;
 }
}
static uintptr_t skeleton_layout_54a310(int i) {
 switch(i) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,obj_subclass);
 case 2:return offsetof(nox_object_t,obj_flags);
 case 3:return offsetof(nox_object_t,x);
 case 4:return offsetof(nox_object_t,y);
 case 5:return offsetof(nox_object_t,field_27);
 case 6:return offsetof(nox_object_t,field_29);
 case 7:return offsetof(nox_object_t,use_data);
 case 8:return sizeof(nox_object_t);
 case 9:return offsetof(nox_modifier_attrs_t,field_16);
 default:return sizeof(nox_modifier_attrs_t);
 }
}
static void* skeleton_bad_54a310(int role) {return (void*)(uintptr_t)(role==-1?0:0x41000000);}
static void* skeleton_name_54a310(int i) {
 static char* names[]={"TestItem","WeaponPower1","Material2","FireEffect","OtherEffect","Skull","ArmBone","LegBone","Sword","WoodenShield","SteelShield"};
 if(i<0)return skeleton_bad_54a310(i);return names[i];
}
static void* skeleton_bone_54a310(void* table,int32_t i) {
 int32_t offset=(int32_t)((uint32_t)i*4u);
 // The original scale/add wraps in PE32 before the pointer load. The test
 // table uses native slots but retains those signed DWORD slot aliases.
 return ((void**)table)[offset/4];
}
static void skeleton_control_54a310(nox_object_t* u,uintptr_t* out,uint32_t* words) {
 out[0]=(uintptr_t)u;out[1]=(uintptr_t)u->use_data;out[2]=(uintptr_t)&u->x;
 memcpy(words,&u->x,8);
}
static nox_object_t* skeleton_new_54a310(char* name) {return nox_xxx_newObjectByTypeID_4E3810(name);}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func skeletonOriginalC54A310(entry int, unit *server.Object, name unsafe.Pointer, mods [4]unsafe.Pointer, charge int32) int32 {
	var cm [4]*C.char
	for i, p := range mods {
		cm[i] = (*C.char)(p)
	}
	return int32(C.skeleton_public_54a310(C.int(entry), asObjectC(unit), (*C.char)(name), &cm[0], C.int(charge)))
}
func skeletonOriginalABI54A310(entry int) int { return int(C.skeleton_abi_54a310(C.int(entry))) }
func skeletonOriginalLayout54A310() (out [11]uintptr) {
	for i := range out {
		out[i] = uintptr(C.skeleton_layout_54a310(C.int(i)))
	}
	return
}
func skeletonOriginalBad54A310(role int) unsafe.Pointer { return C.skeleton_bad_54a310(C.int(role)) }
func skeletonOriginalName54A310(i int) unsafe.Pointer   { return C.skeleton_name_54a310(C.int(i)) }
func skeletonOriginalBone54A310(table unsafe.Pointer, i int32) unsafe.Pointer {
	return C.skeleton_bone_54a310(table, C.int32_t(i))
}
func skeletonOriginalControl54A310(unit *server.Object) (out [3]uintptr, words [2]uint32) {
	var p [3]C.uintptr_t
	var w [2]C.uint32_t
	C.skeleton_control_54a310(asObjectC(unit), &p[0], &w[0])
	for i := range out {
		out[i] = uintptr(p[i])
	}
	for i := range words {
		words[i] = uint32(w[i])
	}
	return
}
func skeletonOriginalNew54A310(name unsafe.Pointer) *server.Object {
	return asObjectS(C.skeleton_new_54a310((*C.char)(name)))
}

var skeletonOriginalRuntimeOverride54A310 *server.SkeletonDeathRuntime54A310
var skeletonOriginalStateOverride54A310 *server.SkeletonDeathState54A310
