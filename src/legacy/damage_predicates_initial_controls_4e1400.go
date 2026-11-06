package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "damage_predicates_raw_4e1400.h"

// No production C WandUseData typedef exists. This pointer-free C payload
// view mirrors the actual named Go record, with all layout facts compared.
typedef struct {
 uint32_t kind;
 uint8_t projectile_name[80];
 uint32_t projectile_type,sound,spell,flags,cooldown,last_used;
 uint8_t charge,max_charge,reserved[2];
 uint32_t progress;
} damage_predicate_wand_view_4e1400;
_Static_assert(sizeof(damage_predicate_wand_view_4e1400)==116,"exact pointer-free Wand payload");
_Static_assert(offsetof(damage_predicate_wand_view_4e1400,flags)==96,"original Wand flags storage");

typedef int (*damage_attack_old_4e1400)(int,uint32_t*);
typedef int32_t (*damage_attack_native_4e1400)(nox_object_t*,nox_object_t*);
typedef int (*damage_friendly_old_4e1400)(int);
typedef int32_t (*damage_friendly_native_4e1400)(nox_object_t*);
static int32_t damage_attack_call_old_4e1400(damage_attack_old_4e1400 fn,void* source,void* weapon) {
 return fn((int)(uintptr_t)source,(uint32_t*)weapon);
}
static int32_t damage_attack_call_native_4e1400(damage_attack_native_4e1400 fn,void* source,void* weapon) {
 return fn(source,weapon);
}
static int32_t damage_friendly_call_old_4e1400(damage_friendly_old_4e1400 fn,void* weapon) {
 return fn((int)(uintptr_t)weapon);
}
static int32_t damage_friendly_call_native_4e1400(damage_friendly_native_4e1400 fn,void* weapon) {
 return fn(weapon);
}
static int damage_predicate_abi_4e1400(int entry) {
 if(!entry)return _Generic(&sub_4E1400,damage_attack_old_4e1400:1,damage_attack_native_4e1400:2);
 return _Generic(&sub_4E1470,damage_friendly_old_4e1400:1,damage_friendly_native_4e1400:2);
}
static int32_t damage_predicate_public_4e1400(int entry,void* source,void* weapon) {
 if(!entry)return _Generic(&sub_4E1400,damage_attack_old_4e1400:damage_attack_call_old_4e1400,damage_attack_native_4e1400:damage_attack_call_native_4e1400)(&sub_4E1400,source,weapon);
 return _Generic(&sub_4E1470,damage_friendly_old_4e1400:damage_friendly_call_old_4e1400,damage_friendly_native_4e1400:damage_friendly_call_native_4e1400)(&sub_4E1470,weapon);
}
static uintptr_t damage_predicate_layout_4e1400(int index) {
 switch(index) {
 case 0:case 2:return offsetof(nox_object_t,obj_class);
 case 1:case 3:return offsetof(nox_object_t,obj_subclass);
 case 4:return offsetof(nox_object_t,use_data);
 case 5:return offsetof(damage_predicate_wand_view_4e1400,flags);
 case 6:return sizeof(void*);
 case 7:return sizeof(nox_object_t);
 case 8:return sizeof(damage_predicate_wand_view_4e1400);
 case 9:return _Alignof(nox_object_t);
 case 10:return _Alignof(damage_predicate_wand_view_4e1400);
 case 11:return sizeof(((nox_object_t*)0)->obj_class);
 case 12:return sizeof(((nox_object_t*)0)->obj_subclass);
 case 13:return sizeof(((nox_object_t*)0)->use_data);
 default:return sizeof(((damage_predicate_wand_view_4e1400*)0)->flags);
 }
}
// ONLY call for independently readable fields. memcpy preserves the real
// offsets and widths without introducing C struct alignment UB on partial
// last-byte and deliberately unaligned fixture objects.
static uintptr_t damage_predicate_initial_read_4e1400(void* base,int kind) {
 const unsigned char* address=(const unsigned char*)base+damage_predicate_layout_4e1400(kind);
 if(kind==1||kind==5){uint8_t value;memcpy(&value,address,1);return value;}
 if(kind==4){uintptr_t value;memcpy(&value,address,sizeof value);return value;}
 uint32_t value;memcpy(&value,address,4);return value;
}
*/
import "C"

import "unsafe"

func damagePredicateLayout4E1400() (out [15]uintptr) {
	for i := range out {
		out[i] = uintptr(C.damage_predicate_layout_4e1400(C.int(i)))
	}
	return
}

func damagePredicateABI4E1400(entry int) int { return int(C.damage_predicate_abi_4e1400(C.int(entry))) }

func damagePredicatePublic4E1400(entry int, source, weapon unsafe.Pointer) int32 {
	return int32(C.damage_predicate_public_4e1400(C.int(entry), source, weapon))
}

func damagePredicateInitialC4E1400(base unsafe.Pointer, kind int) uintptr {
	return uintptr(C.damage_predicate_initial_read_4e1400(base, C.int(kind)))
}
