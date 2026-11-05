package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"

typedef int (*modifier_predicate_old_4e03d0)(int,int,int,int,int,int*);
typedef int32_t (*modifier_predicate_native_4e03d0)(void*,nox_object_t*,uintptr_t,nox_object_t*,uintptr_t,int32_t*);
typedef int (*modifier_scan_old_4e03d0)(int,int,int,int);
typedef int32_t (*modifier_scan_native_4e03d0)(uintptr_t,uintptr_t,nox_object_t*,uintptr_t);
static int32_t modifier_predicate_call_old_4e03d0(modifier_predicate_old_4e03d0 fn,void* effect,const uintptr_t* a,int32_t* out){
 return fn((int)(uintptr_t)effect,(int)a[0],(int)a[1],(int)a[2],(int)a[3],out);
}
static int32_t modifier_predicate_call_native_4e03d0(modifier_predicate_native_4e03d0 fn,void* effect,const uintptr_t* a,int32_t* out){
 return fn(effect,(nox_object_t*)a[0],a[1],(nox_object_t*)a[2],a[3],out);
}
#define MODIFIER_PREDICATE_CALL_4E03D0(fn,e,a,o) _Generic(&(fn),modifier_predicate_old_4e03d0:modifier_predicate_call_old_4e03d0,modifier_predicate_native_4e03d0:modifier_predicate_call_native_4e03d0)(&(fn),e,a,o)
static int32_t modifier_scan_call_old_4e03d0(modifier_scan_old_4e03d0 fn,void* item,const uintptr_t* a){
 return fn((int)a[0],(int)a[1],(int)(uintptr_t)item,(int)a[2]);
}
static int32_t modifier_scan_call_native_4e03d0(modifier_scan_native_4e03d0 fn,void* item,const uintptr_t* a){
 return fn(a[0],a[1],item,a[2]);
}
static int32_t modifier_predicates_public_4e03d0(int which,void* argument,const uintptr_t* a,int32_t* out){
 switch(which){
 case 0:return MODIFIER_PREDICATE_CALL_4E03D0(nox_xxx_inversionEffect_4E03D0,argument,a,out);
 case 1:return MODIFIER_PREDICATE_CALL_4E03D0(nox_xxx_gripEffect_4E0480,argument,a,out);
 default:return _Generic(&nox_xxx_unusedCheckGripEffect_4E03F0,modifier_scan_old_4e03d0:modifier_scan_call_old_4e03d0,modifier_scan_native_4e03d0:modifier_scan_call_native_4e03d0)(&nox_xxx_unusedCheckGripEffect_4E03F0,argument,a);
 }
}
static int modifier_predicates_abi_4e03d0(int which){
 switch(which){
 case 0:return _Generic(&nox_xxx_inversionEffect_4E03D0,modifier_predicate_old_4e03d0:1,modifier_predicate_native_4e03d0:2);
 case 1:return _Generic(&nox_xxx_gripEffect_4E0480,modifier_predicate_old_4e03d0:1,modifier_predicate_native_4e03d0:2);
 default:return _Generic(&nox_xxx_unusedCheckGripEffect_4E03F0,modifier_scan_old_4e03d0:1,modifier_scan_native_4e03d0:2);
 }
}
static void* modifier_predicates_callback_4e03d0(int which){
 return which ? (void*)nox_xxx_gripEffect_4E0480 : (void*)nox_xxx_inversionEffect_4E03D0;
}
static uintptr_t modifier_predicates_layout_4e03d0(int field){
 switch(field){
 case 0:return sizeof(obj_412ae0_t);
 case 1:return offsetof(obj_412ae0_t,field_22);
 case 2:return offsetof(obj_412ae0_t,field_24);
 case 3:return offsetof(obj_412ae0_t,field_1);
 case 4:return sizeof(nox_object_t);
 case 5:return offsetof(nox_object_t,obj_class);
 case 6:return offsetof(nox_object_t,init_data);
 case 7:return sizeof(nox_modifier_attrs_t);
 case 8:return offsetof(nox_modifier_attrs_t,modifiers[0]);
 case 9:return offsetof(nox_modifier_attrs_t,modifiers[1]);
 case 10:return offsetof(nox_modifier_attrs_t,modifiers[2]);
 case 11:return offsetof(nox_modifier_attrs_t,modifiers[3]);
 default:return sizeof(void*);
 }
}
// Initial-memory controls, independent of the original and future engines.
// Call these C getters only for domains independently known to be readable.
static uintptr_t modifier_predicates_initial_read_4e03d0(void* base,int kind,int index){
 switch(kind){
 case 0:return (uint32_t)((obj_412ae0_t*)base)->field_24;
 case 1:return ((nox_object_t*)base)->obj_class;
 case 2:return (uintptr_t)((nox_object_t*)base)->init_data;
 case 3:return (uintptr_t)((nox_modifier_attrs_t*)base)->modifiers[index];
 case 4:return (uintptr_t)((obj_412ae0_t*)base)->field_22;
 default:{uint32_t word;memcpy(&word,base,4);return word;}
 }
}
*/
import "C"

import "unsafe"

func modifierPredicatesLayout4E03D0() (out [13]uintptr) {
	for i := range out {
		out[i] = uintptr(C.modifier_predicates_layout_4e03d0(C.int(i)))
	}
	return
}

func modifierPredicatesABI4E03D0(which int) int {
	return int(C.modifier_predicates_abi_4e03d0(C.int(which)))
}

func modifierPredicatesCallback4E03D0(which int) unsafe.Pointer {
	return C.modifier_predicates_callback_4e03d0(C.int(which))
}

func modifierPredicatesInitialC4E03D0(p unsafe.Pointer, kind, index int) uintptr {
	return uintptr(C.modifier_predicates_initial_read_4e03d0(p, C.int(kind), C.int(index)))
}

func modifierPredicatesPublic4E03D0(which int, p unsafe.Pointer, opaque [4]uint32, output unsafe.Pointer) int32 {
	var args [4]C.uintptr_t
	for i, value := range opaque {
		args[i] = C.uintptr_t(value)
	}
	return int32(C.modifier_predicates_public_4e03d0(C.int(which), p, &args[0], (*C.int32_t)(output)))
}
