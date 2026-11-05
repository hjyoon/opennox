package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4.h"
typedef int (*inversion_inventory_old_4fa4f0)(int,int);
typedef int32_t (*inversion_inventory_native_4fa4f0)(nox_object_t*,nox_object_t*);
static int32_t inversion_inventory_call_old_4fa4f0(inversion_inventory_old_4fa4f0 fn,nox_object_t* root,nox_object_t* target){return fn((int)(uintptr_t)root,(int)(uintptr_t)target);}
static int32_t inversion_inventory_call_native_4fa4f0(inversion_inventory_native_4fa4f0 fn,nox_object_t* root,nox_object_t* target){return fn(root,target);}
static int32_t inversion_inventory_public_4fa4f0(nox_object_t* root,nox_object_t* target){
 return _Generic(&nox_xxx_checkInversionEffect_4FA4F0,inversion_inventory_old_4fa4f0:inversion_inventory_call_old_4fa4f0,inversion_inventory_native_4fa4f0:inversion_inventory_call_native_4fa4f0)(&nox_xxx_checkInversionEffect_4FA4F0,root,target);
}
static int inversion_inventory_abi_4fa4f0(void){return _Generic(&nox_xxx_checkInversionEffect_4FA4F0,inversion_inventory_old_4fa4f0:1,inversion_inventory_native_4fa4f0:2);}
static void* inversion_inventory_pointer_4fa4f0(int parent){return parent?(void*)nox_xxx_findParentChainPlayer_4EC580:(void*)nox_xxx_checkInversionEffect_4FA4F0;}
static uintptr_t inversion_inventory_layout_4fa4f0(int field){
 switch(field){
 case 0:return offsetof(nox_object_t,inv_first_item);
 case 1:return offsetof(nox_object_t,inv_next_item);
 case 2:return offsetof(nox_object_t,owner);
 case 3:return offsetof(nox_object_t,obj_flags);
 default:return offsetof(nox_object_t,obj_class);
 }
}
// Independent typed native controls. The raw checker body remains unchanged.
static uintptr_t inversion_inventory_initial_read_4fa4f0(nox_object_t* root,int kind){
 switch(kind){
 case 6:return (uintptr_t)root->inv_first_item;
 case 7:return (uintptr_t)root->inv_next_item;
 case 8:return (uintptr_t)root->owner;
 case 9:return root->obj_flags;
 default:return ((uint8_t*)&root->obj_class)[0];
 }
}
*/
import "C"

import "unsafe"

func inversionInventoryLayout4FA4F0() (out [5]uintptr) {
	for i := range out {
		out[i] = uintptr(C.inversion_inventory_layout_4fa4f0(C.int(i)))
	}
	return
}

func inversionInventoryABI4FA4F0() int {
	return int(C.inversion_inventory_abi_4fa4f0())
}

func inversionInventoryPointer4FA4F0(parent bool) unsafe.Pointer {
	index := 0
	if parent {
		index = 1
	}
	return C.inversion_inventory_pointer_4fa4f0(C.int(index))
}

func inversionInventoryInitialC4FA4F0(p unsafe.Pointer, kind, index int) uintptr {
	if kind < 6 {
		return modifierPredicatesInitialC4E03D0(p, kind, index)
	}
	return uintptr(C.inversion_inventory_initial_read_4fa4f0((*C.nox_object_t)(p), C.int(kind)))
}

func inversionInventoryParent4FA4F0(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_findParentChainPlayer_4EC580((*C.nox_object_t)(p)))
}

func inversionInventoryPublic4FA4F0(root, target unsafe.Pointer) int32 {
	return int32(C.inversion_inventory_public_4fa4f0((*C.nox_object_t)(root), (*C.nox_object_t)(target)))
}
