package cfixture

/*
#include "monster_spawn_registry_50d780.h"
#include <stddef.h>
#include <string.h>
typedef int(*spawn_int0)(void);
typedef void(*spawn_void0)(void);
typedef unsigned int(*spawn_tick0)(void);
typedef void(*spawn_object1)(nox_object_t*);
typedef nox_monster_spawn_screen*(*spawn_find1)(nox_object_t*);
typedef int(*spawn_place2)(nox_object_t*,float2*);
typedef void(*spawn_visible2)(nox_object_t*,nox_object_t*);
typedef int(*spawn_register2)(nox_object_t*,nox_object_t*);
typedef nox_monster_spawn_link*(*spawn_link1)(nox_monster_spawn_link*);
#define SPAWN_ABI(fn,typ) _Static_assert(_Generic(&(fn),typ:1,default:0),"native pointer ABI: " #fn)
SPAWN_ABI(nox_xxx_allocMonsterRelatedArrays_50D780,spawn_int0);
SPAWN_ABI(sub_50D7E0,spawn_void0);SPAWN_ABI(sub_50D820,spawn_void0);
SPAWN_ABI(sub_50D890,spawn_tick0);SPAWN_ABI(sub_50D8D0,spawn_void0);
SPAWN_ABI(sub_50D960,spawn_int0);SPAWN_ABI(sub_50DE10,spawn_int0);
SPAWN_ABI(sub_50DE60,spawn_find1);SPAWN_ABI(sub_50DE80,spawn_place2);
SPAWN_ABI(sub_50DFB0,spawn_visible2);SPAWN_ABI(sub_50E030,spawn_register2);
SPAWN_ABI(sub_50E110,spawn_link1);SPAWN_ABI(sub_50E140,spawn_object1);
SPAWN_ABI(sub_50E1B0,spawn_link1);SPAWN_ABI(sub_50E1E0,spawn_object1);SPAWN_ABI(sub_50E210,spawn_object1);
#define SPAWN_INT0(fn,typ) do { typ volatile f=(fn);return(uintptr_t)(route?f():(fn)()); } while(0)
#define SPAWN_VOID0(fn) do { spawn_void0 volatile f=(fn);if(route)f();else(fn)();return 0; } while(0)
static uintptr_t spawn_registry_call_C(int route,int entry,void*a,void*b) {
 switch(entry) {
 case 0:SPAWN_INT0(nox_xxx_allocMonsterRelatedArrays_50D780,spawn_int0);
 case 1:SPAWN_VOID0(sub_50D7E0);
 case 2:SPAWN_VOID0(sub_50D820);
 case 3:SPAWN_INT0(sub_50D890,spawn_tick0);
 case 4:SPAWN_VOID0(sub_50D8D0);
 case 5:SPAWN_INT0(sub_50D960,spawn_int0);
 case 6:SPAWN_INT0(sub_50DE10,spawn_int0);
 case 7:{spawn_find1 volatile f=sub_50DE60;return(uintptr_t)(route?f(a):sub_50DE60(a));}
 case 8:{spawn_place2 volatile f=sub_50DE80;return(uintptr_t)(route?f(a,b):sub_50DE80(a,b));}
 case 9:{spawn_visible2 volatile f=sub_50DFB0;if(route)f(a,b);else sub_50DFB0(a,b);return 0;}
 case 10:{spawn_register2 volatile f=sub_50E030;return(uintptr_t)(route?f(a,b):sub_50E030(a,b));}
 case 11:{spawn_link1 volatile f=sub_50E110;return(uintptr_t)(route?f(a):sub_50E110(a));}
 case 12:{spawn_object1 volatile f=sub_50E140;if(route)f(a);else sub_50E140(a);return 0;}
 case 13:{spawn_link1 volatile f=sub_50E1B0;return(uintptr_t)(route?f(a):sub_50E1B0(a));}
 case 14:{spawn_object1 volatile f=sub_50E1E0;if(route)f(a);else sub_50E1E0(a);return 0;}
 case 15:{spawn_object1 volatile f=sub_50E210;if(route)f(a);else sub_50E210(a);return 0;}
 }
 return UINTPTR_MAX;
}
static uintptr_t spawn_registry_read_C(void*owner,size_t offset) {
 uintptr_t value;memcpy(&value,(unsigned char*)owner+offset,sizeof(value));return value;
}
static uintptr_t spawn_registry_data_C(nox_object_t*object) { return(uintptr_t)object->data_update; }
static void spawn_registry_layout_C(uintptr_t*v) {
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,data_update);
 v[3]=sizeof(nox_monster_spawn_link);v[4]=offsetof(nox_monster_spawn_link,older);v[5]=offsetof(nox_monster_spawn_link,newer);
 v[6]=sizeof(nox_monster_spawn_screen);v[7]=offsetof(nox_monster_spawn_screen,average);v[8]=offsetof(nox_monster_spawn_screen,mask);v[9]=offsetof(nox_monster_spawn_screen,distances);v[10]=offsetof(nox_monster_spawn_screen,older);v[11]=offsetof(nox_monster_spawn_screen,newer);
}
*/
import "C"

import "unsafe"

func MonsterSpawnRegistryNativeCall(route, entry int, a, b unsafe.Pointer) uintptr {
	return uintptr(C.spawn_registry_call_C(C.int(route), C.int(entry), a, b))
}
func MonsterSpawnRegistryNativeRead(owner unsafe.Pointer, offset uintptr) uintptr {
	return uintptr(C.spawn_registry_read_C(owner, C.size_t(offset)))
}
func MonsterSpawnRegistryNativeObjectData(owner unsafe.Pointer) uintptr {
	return uintptr(C.spawn_registry_data_C((*C.nox_object_t)(owner)))
}
func MonsterSpawnRegistryNativeLayout() (out [12]uintptr) {
	C.spawn_registry_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}
