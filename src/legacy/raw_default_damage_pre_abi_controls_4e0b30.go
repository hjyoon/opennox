package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "unit_buff_test_4ff350.h"

typedef int (*raw_parent_old_4e0b30)(int,int,int,int,int);
typedef int32_t (*raw_parent_native_4e0b30)(nox_object_t*,nox_object_t*,nox_object_t*,int32_t,int32_t);
static int32_t raw_parent_call_old_4e0b30(raw_parent_old_4e0b30 fn,void* target,void* source,void* weapon,int32_t damage,int32_t typ) {
 return fn((int)(uintptr_t)target,(int)(uintptr_t)source,(int)(uintptr_t)weapon,damage,typ);
}
static int32_t raw_parent_call_native_4e0b30(raw_parent_native_4e0b30 fn,void* target,void* source,void* weapon,int32_t damage,int32_t typ) {
 return fn(target,source,weapon,damage,typ);
}
static int raw_parent_abi_4e0b30(void) {
 return _Generic(&nox_xxx_damageDefaultProc_4E0B30,raw_parent_old_4e0b30:1,raw_parent_native_4e0b30:2);
}
static uintptr_t raw_parent_entry_4e0b30(void) {return (uintptr_t)&nox_xxx_damageDefaultProc_4E0B30;}
static int32_t raw_parent_public_4e0b30(void* target,void* source,void* weapon,int32_t damage,int32_t typ) {
 return _Generic(&nox_xxx_damageDefaultProc_4E0B30,
  raw_parent_old_4e0b30:raw_parent_call_old_4e0b30,
  raw_parent_native_4e0b30:raw_parent_call_native_4e0b30)
  (&nox_xxx_damageDefaultProc_4E0B30,target,source,weapon,damage,typ);
}
// Transport-only controls. These are explicit echo callbacks, not an
// invocation or implementation of the original raw damage parent. Observe
// all three losses separately without inventing three whole-function faults.
static _Thread_local uintptr_t raw_parent_echo_words_4e0b30[5];
static int raw_parent_echo_old_4e0b30(int target,int source,int weapon,int damage,int typ) {
 raw_parent_echo_words_4e0b30[0]=(uint32_t)target;
 raw_parent_echo_words_4e0b30[1]=(uint32_t)source;
 raw_parent_echo_words_4e0b30[2]=(uint32_t)weapon;
 raw_parent_echo_words_4e0b30[3]=(uint32_t)damage;
 raw_parent_echo_words_4e0b30[4]=(uint32_t)typ;
 return 1;
}
static int32_t raw_parent_echo_native_4e0b30(nox_object_t* target,nox_object_t* source,nox_object_t* weapon,int32_t damage,int32_t typ) {
 raw_parent_echo_words_4e0b30[0]=(uintptr_t)target;
 raw_parent_echo_words_4e0b30[1]=(uintptr_t)source;
 raw_parent_echo_words_4e0b30[2]=(uintptr_t)weapon;
 raw_parent_echo_words_4e0b30[3]=(uint32_t)damage;
 raw_parent_echo_words_4e0b30[4]=(uint32_t)typ;
 return 1;
}
static void raw_parent_echo_4e0b30(int native,void* target,void* source,void* weapon,int32_t damage,int32_t typ,uintptr_t* out) {
 if(native)raw_parent_call_native_4e0b30(&raw_parent_echo_native_4e0b30,target,source,weapon,damage,typ);
 else raw_parent_call_old_4e0b30(&raw_parent_echo_old_4e0b30,target,source,weapon,damage,typ);
 memcpy(out,raw_parent_echo_words_4e0b30,5*sizeof(uintptr_t));
}
static uintptr_t raw_parent_control_layout_4e0b30(int i) {
 switch(i) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,obj_subclass);
 case 2:return offsetof(nox_object_t,obj_flags);
 case 3:return offsetof(nox_object_t,buffs);
 case 4:return offsetof(nox_object_t,obj_130);
 case 5:return offsetof(nox_object_t,field_131);
 case 6:return offsetof(nox_object_t,field_132);
 case 7:return offsetof(nox_object_t,field_134);
 case 8:return offsetof(nox_object_t,owner);
 case 9:return offsetof(nox_object_t,func_damage);
 case 10:return offsetof(nox_object_t,func_damage_sound);
 case 11:return offsetof(nox_object_t,data_update);
 case 12:return sizeof(nox_object_t);
 case 13:return sizeof(void*);
 case 14:return sizeof(((nox_object_t*)0)->func_damage);
 default:return sizeof(((nox_object_t*)0)->data_update);
 }
}
static int32_t raw_parent_direct_buff_control_4e0b30(void* target,int32_t buff) {
 return nox_xxx_testUnitBuffs_4FF350(target,buff);
}
*/
import "C"

import "unsafe"

func rawParentABI4E0B30() int       { return int(C.raw_parent_abi_4e0b30()) }
func rawParentEntry4E0B30() uintptr { return uintptr(C.raw_parent_entry_4e0b30()) }
func rawParentPublic4E0B30(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	return int32(C.raw_parent_public_4e0b30(target, source, weapon, C.int32_t(damage), C.int32_t(typ)))
}
func rawParentControlLayout4E0B30() (out [16]uintptr) {
	for i := range out {
		out[i] = uintptr(C.raw_parent_control_layout_4e0b30(C.int(i)))
	}
	return
}
func rawParentEcho4E0B30(native bool, target, source, weapon unsafe.Pointer, damage, typ int32) (out [5]uintptr) {
	mode := 0
	if native {
		mode = 1
	}
	C.raw_parent_echo_4e0b30(C.int(mode), target, source, weapon, C.int32_t(damage), C.int32_t(typ), (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}
func rawParentDirectBuff4E0B30(target unsafe.Pointer, buff int32) int32 {
	return int32(C.raw_parent_direct_buff_control_4e0b30(target, C.int32_t(buff)))
}
