package legacy

/*
#include <stdint.h>
#include "damage_predicates_raw_4e1400.h"
typedef int (*damage_raw_attack_old_4e1400)(int,uint32_t*);
typedef int32_t (*damage_raw_attack_native_4e1400)(nox_object_t*,nox_object_t*);
typedef int (*damage_raw_friendly_old_4e1400)(int);
typedef int32_t (*damage_raw_friendly_native_4e1400)(nox_object_t*);
typedef struct {uintptr_t function,args[2];int32_t entry,argc,phase,result;} damage_raw_trace_4e1400;
static int damage_raw_trace_enabled_4e1400,damage_raw_trace_count_4e1400;
static damage_raw_trace_4e1400 damage_raw_trace_data_4e1400[2];
static void damage_raw_trace_begin_4e1400(void){damage_raw_trace_count_4e1400=0;damage_raw_trace_enabled_4e1400=1;}
static void damage_raw_trace_end_4e1400(void){
 damage_raw_trace_enabled_4e1400=0;damage_raw_trace_count_4e1400=0;
 for(int i=0;i<2;i++)damage_raw_trace_data_4e1400[i]=(damage_raw_trace_4e1400){0};
}
static void damage_raw_trace_append_4e1400(int entry,uintptr_t fn,void* source,void* weapon,int phase,int32_t result){
 if(!damage_raw_trace_enabled_4e1400)return;
 int i=damage_raw_trace_count_4e1400++;if(i>=2)return;
 damage_raw_trace_data_4e1400[i]=(damage_raw_trace_4e1400){fn,
  {(uintptr_t)(entry?weapon:source),entry?0:(uintptr_t)weapon},entry,entry?1:2,phase,result};
}
static int32_t damage_raw_attack_call_old_4e1400(damage_raw_attack_old_4e1400 fn,void* source,void* weapon){return fn((int)(uintptr_t)source,(uint32_t*)weapon);}
static int32_t damage_raw_attack_call_native_4e1400(damage_raw_attack_native_4e1400 fn,void* source,void* weapon){return fn(source,weapon);}
static int32_t damage_raw_friendly_call_old_4e1400(damage_raw_friendly_old_4e1400 fn,void* weapon){return fn((int)(uintptr_t)weapon);}
static int32_t damage_raw_friendly_call_native_4e1400(damage_raw_friendly_native_4e1400 fn,void* weapon){return fn(weapon);}
static uintptr_t damage_raw_root_4e1400(int entry){return entry?(uintptr_t)&sub_4E1470:(uintptr_t)&sub_4E1400;}
// Go passes void* without casting an unaligned/partial base to a whole
// pointer-bearing C object. Typed actual root calls remain inside C.
static int32_t damage_raw_public_4e1400(int entry,void* source,void* weapon){
 uintptr_t fn=damage_raw_root_4e1400(entry);int32_t result;
 damage_raw_trace_append_4e1400(entry,fn,source,weapon,0,0);
 if(!entry)result=_Generic(&sub_4E1400,damage_raw_attack_old_4e1400:damage_raw_attack_call_old_4e1400,damage_raw_attack_native_4e1400:damage_raw_attack_call_native_4e1400)(&sub_4E1400,source,weapon);
 else result=_Generic(&sub_4E1470,damage_raw_friendly_old_4e1400:damage_raw_friendly_call_old_4e1400,damage_raw_friendly_native_4e1400:damage_raw_friendly_call_native_4e1400)(&sub_4E1470,weapon);
 damage_raw_trace_append_4e1400(entry,fn,source,weapon,1,result);return result;
}
static int32_t damage_raw_attack_stub_4e1400(nox_object_t* source,nox_object_t* weapon){(void)source;(void)weapon;return -INT32_C(2023406815);}
static int32_t damage_raw_friendly_stub_4e1400(nox_object_t* weapon){(void)weapon;return -INT32_C(19088743);}
static uintptr_t damage_raw_stub_pointer_4e1400(int entry){return entry?(uintptr_t)&damage_raw_friendly_stub_4e1400:(uintptr_t)&damage_raw_attack_stub_4e1400;}
static int32_t damage_raw_transport_4e1400(int entry,void* source,void* weapon){
 uintptr_t fn=damage_raw_stub_pointer_4e1400(entry);int32_t result;
 damage_raw_trace_append_4e1400(entry,fn,source,weapon,0,0);
 if(!entry){damage_raw_attack_native_4e1400 call=damage_raw_attack_stub_4e1400;result=call(source,weapon);}
 else{damage_raw_friendly_native_4e1400 call=damage_raw_friendly_stub_4e1400;result=call(weapon);}
 damage_raw_trace_append_4e1400(entry,fn,source,weapon,1,result);return result;
}
static int damage_raw_trace_size_4e1400(void){return damage_raw_trace_count_4e1400;}
static damage_raw_trace_4e1400 damage_raw_trace_get_4e1400(int i){return damage_raw_trace_data_4e1400[i];}
static int32_t damage_raw_return_4e1400(int32_t value){return value;}
// Opaque bit transport only: no field/pointee dereference and no C memory fault.
static void* damage_raw_pointer_from_bits_4e1400(uintptr_t bits){return (void*)bits;}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type damagePredicateRawValue4E1400 struct {
	Pointer unsafe.Pointer
	Word    uint32
}

// These hooks are nil in ordinary execution. There is no predicate admission,
// class mask, branch result or gameplay algorithm in this transport file.
var damagePredicateRawObserve4E1400 func(complete bool, base unsafe.Pointer, kind int, value damagePredicateRawValue4E1400)

func damagePredicateRawOffset4E1400(kind int) uintptr {
	var object server.Object
	var use server.UseDataPtr
	var wand server.WandUseData
	switch kind {
	case 0, 2:
		return unsafe.Offsetof(object.ObjClass)
	case 1, 3:
		return unsafe.Offsetof(object.ObjSubClass)
	case 4:
		return unsafe.Offsetof(object.UseData) + unsafe.Offsetof(use.Ptr)
	case 5:
		return unsafe.Offsetof(wand.Flags)
	default:
		panic("undeclared raw damage field")
	}
}

func damagePredicateRawRead4E1400(base unsafe.Pointer, kind int) (value damagePredicateRawValue4E1400) {
	if damagePredicateRawObserve4E1400 != nil {
		damagePredicateRawObserve4E1400(false, base, kind, value)
	}
	offset := damagePredicateRawOffset4E1400(kind)
	if base == nil {
		// An actual nil read keeps the recoverable field-fault domain. Avoid
		// fatal checkptr arithmetic from unsafe.Add(nil, namedOffset).
		switch kind {
		case 1, 5:
			value.Word = uint32(*(*uint8)(base))
		case 4:
			value.Pointer = *(*unsafe.Pointer)(base)
		default:
			value.Word = *(*uint32)(base)
		}
	} else {
		address := unsafe.Add(base, offset)
		switch kind {
		case 1, 5:
			value.Word = uint32(*(*uint8)(address))
		case 4:
			if uintptr(address)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
				// Actual aligned Go pointer storage retains GC pointer semantics.
				value.Pointer = *(*unsafe.Pointer)(address)
			} else {
				// Unaligned storage is a C-owned memory domain. A Go-owned
				// typed pointer field cannot legally have this alignment.
				var bits uintptr
				if unsafe.Sizeof(uintptr(0)) == 8 {
					bits = uintptr(*(*uint64)(address))
				} else {
					bits = uintptr(*(*uint32)(address))
				}
				value.Pointer = C.damage_raw_pointer_from_bits_4e1400(C.uintptr_t(bits))
				runtime.KeepAlive(base)
			}
		default:
			value.Word = *(*uint32)(address)
		}
	}
	if damagePredicateRawObserve4E1400 != nil {
		damagePredicateRawObserve4E1400(true, base, kind, value)
	}
	runtime.KeepAlive(base)
	return
}

type damagePredicateRawTrace4E1400 struct {
	Function           uintptr
	Args               [2]uintptr
	Entry, Argc, Phase int
	Result             int32
}

func damagePredicateRawTracing4E1400(enabled bool) {
	if enabled {
		C.damage_raw_trace_begin_4e1400()
	} else {
		C.damage_raw_trace_end_4e1400()
	}
}

func damagePredicateRawTraces4E1400() (out []damagePredicateRawTrace4E1400) {
	n := int(C.damage_raw_trace_size_4e1400())
	if n < 0 || n > 2 {
		panic("raw damage C trace overflow")
	}
	for i := 0; i < n; i++ {
		r := C.damage_raw_trace_get_4e1400(C.int(i))
		out = append(out, damagePredicateRawTrace4E1400{uintptr(r.function), [2]uintptr{uintptr(r.args[0]), uintptr(r.args[1])}, int(r.entry), int(r.argc), int(r.phase), int32(r.result)})
	}
	return
}

func damagePredicateRawRootPointer4E1400(entry int) uintptr {
	return uintptr(C.damage_raw_root_4e1400(C.int(entry)))
}

func damagePredicateRawStubPointer4E1400(entry int) uintptr {
	return uintptr(C.damage_raw_stub_pointer_4e1400(C.int(entry)))
}

func damagePredicateRawTransport4E1400(entry int, source, weapon unsafe.Pointer) int32 {
	return int32(C.damage_raw_transport_4e1400(C.int(entry), source, weapon))
}

func damagePredicateRawReturn4E1400(value int32) int32 {
	return int32(C.damage_raw_return_4e1400(C.int32_t(value)))
}

// Both remain nil until separately frozen raw engines are installed. The
// future public exports must call these exact respective engine functions.
var damagePredicateRawAttackEngine4E1400 func(source, weapon unsafe.Pointer) int32
var damagePredicateRawFriendlyEngine4E1470 func(weapon unsafe.Pointer) int32
var damagePredicateRawObserveRoot4E1400 func(complete bool, entry int, args [2]unsafe.Pointer, result int32)

func damagePredicateRawArgs4E1400(entry int, source, weapon unsafe.Pointer) [2]unsafe.Pointer {
	if entry == 0 {
		return [2]unsafe.Pointer{source, weapon}
	}
	if entry == 1 {
		return [2]unsafe.Pointer{weapon, nil}
	}
	panic("undeclared raw damage entry")
}

func damagePredicateRawInvoke4E1400(public bool, entry int, source, weapon unsafe.Pointer) (result int32) {
	args := damagePredicateRawArgs4E1400(entry, source, weapon)
	if damagePredicateRawObserveRoot4E1400 != nil {
		damagePredicateRawObserveRoot4E1400(false, entry, args, 0)
	}
	if public {
		result = int32(C.damage_raw_public_4e1400(C.int(entry), source, weapon))
	} else if entry == 0 {
		result = damagePredicateRawAttackEngine4E1400(source, weapon)
	} else {
		result = damagePredicateRawFriendlyEngine4E1470(weapon)
	}
	if damagePredicateRawObserveRoot4E1400 != nil {
		damagePredicateRawObserveRoot4E1400(true, entry, args, result)
	}
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	return
}
