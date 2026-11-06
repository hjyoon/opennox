//go:build darwin && arm64

package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"

// Draft initial-memory probes only: no parent/helper algorithm is invoked.
_Static_assert(sizeof(uintptr_t)==sizeof(void*), "uintptr width");
_Static_assert(sizeof(((nox_object_t*)0)->func_damage)==sizeof(uintptr_t), "callback width");
// Explicit pointer-free fixture view, not a claimed production C typedef.
typedef struct {
 uint32_t kind;
 uint8_t projectile_name[80];
 uint32_t projectile_type,sound,spell,flags,cooldown,last_used;
 uint8_t charge,max_charge,reserved[2];
 uint32_t progress;
} raw_parent_initial_wand_view_4e0b30;
_Static_assert(sizeof(raw_parent_initial_wand_view_4e0b30)==116,"Wand payload width");
static uintptr_t raw_parent_initial_c_offset_4e0b30(int id) {
 switch(id) {
 case 0:return offsetof(nox_object_t,typ_ind);
 case 1:return offsetof(nox_object_t,obj_class);
 case 2:return offsetof(nox_object_t,obj_subclass);
 case 3:return offsetof(nox_object_t,obj_flags);
 case 4:return offsetof(nox_object_t,x);
 case 5:return offsetof(nox_object_t,y);
 case 6:return offsetof(nox_object_t,prev_x);
 case 7:return offsetof(nox_object_t,prev_y);
 case 8:return offsetof(nox_object_t,owner);
 case 9:return offsetof(nox_object_t,obj_130);
 case 10:return offsetof(nox_object_t,field_131);
 case 11:return offsetof(nox_object_t,field_132);
 case 12:return offsetof(nox_object_t,field_133);
 case 13:return offsetof(nox_object_t,field_134);
 case 14:return offsetof(nox_object_t,health_data);
 case 15:return offsetof(nox_object_t,func_damage);
 case 16:return offsetof(nox_object_t,func_damage_sound);
 case 17:return offsetof(nox_object_t,use_data);
 case 18:return offsetof(nox_object_t,data_update);
 case 19:return offsetof(nox_player_update_data_t,state);
 case 20:return offsetof(nox_player_update_data_t,field_40_0);
 // There is no named C monster-update struct in current defs.h. An absent
 // offsetof must remain absent; Go's named MonsterUpdateData supplies it.
 case 21:case 22:case 23:case 24:return UINTPTR_MAX;
 case 25:return offsetof(raw_parent_initial_wand_view_4e0b30,flags);
 case 26:return 32; // external fixed-byte soundset scalar ID
 default:return UINTPTR_MAX;
 }
}
static uintptr_t raw_parent_initial_c_size_4e0b30(int id) {
 switch(id) {
 case 0:return sizeof(nox_object_t);
 case 1:return sizeof(void*);
 case 2:return sizeof(((nox_object_t*)0)->func_damage);
 case 3:return sizeof(nox_player_update_data_t);
 default:return 0;
 }
}
// CALL ONLY AFTER layout and mapped-span gates establish the entire access
// is readable. Never use C memcpy for an in-process protected-field probe.
static uintptr_t raw_parent_initial_c_bits_4e0b30(void* base,uintptr_t offset,int width) {
 uintptr_t value=0;
 size_t n=width==0?sizeof(value):(size_t)width;
 if(n>sizeof(value))abort();
 memcpy(&value,(unsigned char*)base+offset,n);
 return value;
}
static void raw_parent_initial_set_pointer_4e0b30(void* base,uintptr_t offset,void* value) {
 memcpy((unsigned char*)base+offset,&value,sizeof(value));
}
static void* raw_parent_initial_foreign_bits_pointer_4e0b30(uintptr_t bits) {
 return (void*)bits; // pure cast, no dereference; C-owned/foreign memory only
}
static void* raw_parent_initial_calloc_4e0b30(size_t bytes) {return calloc(1,bytes);}
static void raw_parent_initial_free_4e0b30(void* p) {free(p);}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Field IDs are frozen by the independent initial descriptor's27-row table.
func rawParentInitialGoOffsets4E0B30() [27]uintptr {
	return [27]uintptr{
		unsafe.Offsetof(server.Object{}.TypeInd),
		unsafe.Offsetof(server.Object{}.ObjClass),
		unsafe.Offsetof(server.Object{}.ObjSubClass),
		unsafe.Offsetof(server.Object{}.ObjFlags),
		unsafe.Offsetof(server.Object{}.PosVec),
		unsafe.Offsetof(server.Object{}.PosVec) + 4,
		unsafe.Offsetof(server.Object{}.PrevPos),
		unsafe.Offsetof(server.Object{}.PrevPos) + 4,
		unsafe.Offsetof(server.Object{}.ObjOwner),
		unsafe.Offsetof(server.Object{}.Obj130),
		unsafe.Offsetof(server.Object{}.Field131),
		unsafe.Offsetof(server.Object{}.Pos132),
		unsafe.Offsetof(server.Object{}.Pos132) + 4,
		unsafe.Offsetof(server.Object{}.Frame134),
		unsafe.Offsetof(server.Object{}.HealthData),
		unsafe.Offsetof(server.Object{}.Damage),
		unsafe.Offsetof(server.Object{}.DamageSound),
		unsafe.Offsetof(server.Object{}.UseData),
		unsafe.Offsetof(server.Object{}.UpdateData),
		unsafe.Offsetof(server.PlayerUpdateData{}.State),
		unsafe.Offsetof(server.PlayerUpdateData{}.Field40_0),
		unsafe.Offsetof(server.MonsterUpdateData{}.StatusFlags),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field523_2),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field546),
		unsafe.Offsetof(server.MonsterUpdateData{}.Field547),
		unsafe.Offsetof(server.WandUseData{}.Flags),
		32,
	}
}

func rawParentInitialCOffsets4E0B30() (out [27]uintptr) {
	for i := range out {
		out[i] = uintptr(C.raw_parent_initial_c_offset_4e0b30(C.int(i)))
	}
	return
}

func rawParentInitialCSizes4E0B30() (out [4]uintptr) {
	for i := range out {
		out[i] = uintptr(C.raw_parent_initial_c_size_4e0b30(C.int(i)))
	}
	return
}

// Foreign C-owned storage may be unaligned. Primitive pointer-free loads expose
// exact width to checkptr and produce recoverable Go faults on protected input.
// A nil base faults before unsafe.Add; a safe-return nil guard is incorrect.
// This is an initial probe, not a replacement for future actual typed Goheap
// pointer loads. Never convert a Go heap pointer to uintptr and back through it.
func rawParentInitialGoBits4E0B30(base unsafe.Pointer, offset uintptr, width int) uintptr {
	if base == nil {
		return uintptr(*(*byte)(base))
	}
	address := unsafe.Add(base, offset)
	var value uintptr
	switch width {
	case 0:
		value = *(*uintptr)(address)
	case 1:
		value = uintptr(*(*uint8)(address))
	case 2:
		value = uintptr(*(*uint16)(address))
	case 4:
		value = uintptr(*(*uint32)(address))
	default:
		panic("unsupported initial probe width")
	}
	runtime.KeepAlive(base)
	return value
}

func rawParentInitialCBits4E0B30(base unsafe.Pointer, offset uintptr, width int) uintptr {
	return uintptr(C.raw_parent_initial_c_bits_4e0b30(base, C.uintptr_t(offset), C.int(width)))
}

func rawParentInitialSetPointer4E0B30(base unsafe.Pointer, offset uintptr, value unsafe.Pointer) {
	C.raw_parent_initial_set_pointer_4e0b30(base, C.uintptr_t(offset), value)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
}

func rawParentInitialForeignBitsPointer4E0B30(bits uintptr) unsafe.Pointer {
	return C.raw_parent_initial_foreign_bits_pointer_4e0b30(C.uintptr_t(bits))
}

func rawParentInitialAlloc4E0B30(bytes uintptr) unsafe.Pointer {
	return C.raw_parent_initial_calloc_4e0b30(C.size_t(bytes))
}

func rawParentInitialFree4E0B30(base unsafe.Pointer) {
	C.raw_parent_initial_free_4e0b30(base)
}

// Typed aligned-object control exposes the real Go pointer members, rather
// than a second raw uint32 transport. Guarded/unaligned probes use the separate
// primitive loader above and do not gain typed-Goheap/default-service credit.
func rawParentInitialGoNamedPointer4E0B30(base unsafe.Pointer, id int) uintptr {
	o := (*server.Object)(base)
	var value unsafe.Pointer
	switch id {
	case 8:
		value = unsafe.Pointer(o.ObjOwner)
	case 9:
		value = unsafe.Pointer(o.Obj130)
	case 14:
		value = unsafe.Pointer(o.HealthData)
	case 15:
		value = o.Damage
	case 16:
		value = o.DamageSound
	case 17:
		value = o.UseData.Ptr
	case 18:
		value = o.UpdateData
	default:
		panic("non-pointer field in named initial control")
	}
	runtime.KeepAlive(base)
	return uintptr(value)
}
