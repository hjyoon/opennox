package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"

static int simple_polyp_old_int_54cae0(int(*fn)(int),nox_object_t* u) {return fn((int)(uintptr_t)u);}
static int simple_polyp_native_int_54cae0(int(*fn)(nox_object_t*),nox_object_t* u) {return fn(u);}
static void simple_polyp_old_void_54cae0(void(*fn)(int),nox_object_t* u) {fn((int)(uintptr_t)u);}
static void simple_polyp_native_void_54cae0(void(*fn)(nox_object_t*),nox_object_t* u) {fn(u);}
static uint32_t simple_polyp_public_54cae0(int entry,nox_object_t* u) {
 if(entry==0)return (uint32_t)_Generic(&nox_xxx_dieImpEgg_54CAE0,
 int(*)(int):simple_polyp_old_int_54cae0,int(*)(nox_object_t*):simple_polyp_native_int_54cae0)(&nox_xxx_dieImpEgg_54CAE0,u);
 if(entry==1)_Generic(&nox_xxx_diePolyp_54CB10,
 void(*)(int):simple_polyp_old_void_54cae0,void(*)(nox_object_t*):simple_polyp_native_void_54cae0)(&nox_xxx_diePolyp_54CB10,u);
 else _Generic(&nox_xxx_diePotion_54CBB0,
 void(*)(int):simple_polyp_old_void_54cae0,void(*)(nox_object_t*):simple_polyp_native_void_54cae0)(&nox_xxx_diePotion_54CBB0,u);
 return 0;
}
static int simple_polyp_abi_54cae0(int entry) {
 if(entry==0)return _Generic(&nox_xxx_dieImpEgg_54CAE0,int(*)(int):32,int(*)(nox_object_t*):64);
 if(entry==1)return _Generic(&nox_xxx_diePolyp_54CB10,void(*)(int):32,void(*)(nox_object_t*):64);
 return _Generic(&nox_xxx_diePotion_54CBB0,void(*)(int):32,void(*)(nox_object_t*):64);
}
static uintptr_t simple_polyp_layout_54cae0(int index) {
 switch(index) {
 case 0:return offsetof(nox_object_t,obj_flags);
 case 1:return offsetof(nox_object_t,x);
 case 2:return offsetof(nox_object_t,y);
 case 3:return offsetof(nox_object_t,data_update);
 default:return sizeof(nox_object_t);
 }
}
static void simple_polyp_control_54cae0(nox_object_t* u,uintptr_t* pointers,uint32_t* words) {
 pointers[0]=(uintptr_t)u;pointers[1]=(uintptr_t)u->data_update;
 pointers[2]=(uintptr_t)&u->x;pointers[3]=(uintptr_t)&u->y;
 words[0]=u->obj_flags;memcpy(words+1,&u->x,8);
 words[3]=u->data_update?*(uint32_t*)u->data_update:0;
}
static uint32_t simple_polyp_type_54cae0(void) {return (uint32_t)nox_xxx_getNameId_4E3AA0("ToxicCloud");}
static nox_object_t* simple_polyp_new_54cae0(uint32_t id) {return nox_xxx_newObjectWithTypeInd_4E3450((int32_t)id);}
static void simple_polyp_audio_control_54cae0(uint32_t id,nox_object_t* u) {nox_xxx_aud_501960((int32_t)id,u,0,0);}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func simplePolypOriginalC54CAE0(entry int, source *server.Object) uint32 {
	return uint32(C.simple_polyp_public_54cae0(C.int(entry), asObjectC(source)))
}

func simplePolypOriginalABI54CAE0(entry int) int {
	return int(C.simple_polyp_abi_54cae0(C.int(entry)))
}

func simplePolypOriginalLayout54CAE0() (out [5]uintptr) {
	for i := range out {
		out[i] = uintptr(C.simple_polyp_layout_54cae0(C.int(i)))
	}
	return
}

func simplePolypOriginalControl54CAE0(source *server.Object) (pointers [4]uintptr, words [4]uint32) {
	var p [4]C.uintptr_t
	var w [4]C.uint32_t
	C.simple_polyp_control_54cae0(asObjectC(source), &p[0], &w[0])
	for i := range p {
		pointers[i], words[i] = uintptr(p[i]), uint32(w[i])
	}
	return
}

func simplePolypOriginalType54CAE0() uint32 { return uint32(C.simple_polyp_type_54cae0()) }
func simplePolypOriginalNew54CAE0(id uint32) *server.Object {
	return asObjectS(C.simple_polyp_new_54cae0(C.uint32_t(id)))
}
func simplePolypOriginalAudio54CAE0(id uint32, source *server.Object) {
	C.simple_polyp_audio_control_54cae0(C.uint32_t(id), (*C.nox_object_t)(unsafe.Pointer(source)))
}

var simplePolypOriginalSimpleRuntimeOverride54CAE0 *server.SimpleObjectDeathRuntime54CAE0
var simplePolypOriginalRuntimeOverride54CB10 *server.PolypDeathRawRuntime54CB10
var simplePolypOriginalStateOverride54CB10 *server.PolypDeathState54CB10
