package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5.h"
static int golem_old_call_549e90(int(*fn)(int),nox_object_t* u) {return fn((int)(uintptr_t)u);}
static int golem_native_call_549e90(int(*fn)(nox_object_t*),nox_object_t* u) {return fn(u);}
#define GOLEM_CALL(fn,u) _Generic(&(fn),int(*)(int):golem_old_call_549e90,int(*)(nox_object_t*):golem_native_call_549e90)(&(fn),u)
static int golem_public_549e90(int mech,nox_object_t* u) {
 if(mech)return GOLEM_CALL(sub_549E90,u);return GOLEM_CALL(sub_549FA0,u);
}
static int golem_abi_549e90(int mech) {
 if(mech)return _Generic(&sub_549E90,int(*)(int):32,int(*)(nox_object_t*):64);
 return _Generic(&sub_549FA0,int(*)(int):32,int(*)(nox_object_t*):64);
}
static uintptr_t golem_layout_549e90(int i) {
 switch(i) {
 case 0:return offsetof(nox_object_t,obj_flags);
 case 1:return offsetof(nox_object_t,x);
 case 2:return offsetof(nox_object_t,y);
 case 3:return offsetof(nox_object_t,field_27);
 case 4:return offsetof(nox_object_t,field_29);
 case 5:return offsetof(nox_object_t,use_data);
 case 6:return offsetof(nox_object_t,data_update);
 default:return sizeof(nox_object_t);
 }
}
static void* golem_bad_549e90(int role) {return (void*)(uintptr_t)(role==-1?0:0x41000000);}
static void* golem_name_549e90(int i) {
 static char* names[]={"MechGear","MechHead","MechArm","MechLeg","BigRock","MediumRock","SmallRock",""};
 if(i<0)return golem_bad_549e90(i);return names[i];
}
// Return an address, not a C load, so deliberate malformed table reads occur
// in Go under SetPanicOnFault rather than becoming unrecoverable cgo faults.
static void* golem_slot_address_549e90(void* table,int32_t i) {
 int32_t offset=(int32_t)((uint32_t)i*4u);
 return (void*)((char**)table+offset/4);
}
static void* golem_byte_address_549e90(void* table,int32_t i) {return (char*)table+i;}
static void* golem_mass_bytes_549e90(void) {static const unsigned char data[]={80,169,84,0,0,0,0,0,0,0,0,0,0,0,240,65,1,3,5,3,5,5,3,5,5,5,5,5,12,0,0,0,2,0,0,0,79,103,114,101,83,116,114,105,107,101,0,0};return (void*)data;}
static void* golem_bad_name_549e90(void) {return (void*)(uintptr_t)0x52676942;}
static void golem_control_549e90(nox_object_t* u,uintptr_t* out,uint32_t* words) {
 out[0]=(uintptr_t)u;out[1]=(uintptr_t)u->use_data;out[2]=(uintptr_t)u->data_update;out[3]=(uintptr_t)&u->x;
 memcpy(words,&u->x,8);words[2]=u->obj_flags;memcpy(words+3,&u->field_27,4);memcpy(words+4,&u->field_29,4);
}
static nox_object_t* golem_new_549e90(char* name) {return nox_xxx_newObjectByTypeID_4E3810(name);}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func golemOriginalC549E90(mech bool, unit *server.Object) int32 {
	n := C.int(0)
	if mech {
		n = 1
	}
	return int32(C.golem_public_549e90(n, asObjectC(unit)))
}
func golemOriginalABI549E90(mech bool) int {
	n := C.int(0)
	if mech {
		n = 1
	}
	return int(C.golem_abi_549e90(n))
}
func golemOriginalLayout549E90() (out [8]uintptr) {
	for i := range out {
		out[i] = uintptr(C.golem_layout_549e90(C.int(i)))
	}
	return
}
func golemOriginalBad549E90(role int) unsafe.Pointer { return C.golem_bad_549e90(C.int(role)) }
func golemOriginalName549E90(i int) unsafe.Pointer   { return C.golem_name_549e90(C.int(i)) }
func golemOriginalSlotAddress549E90(table unsafe.Pointer, i int32) unsafe.Pointer {
	return C.golem_slot_address_549e90(table, C.int32_t(i))
}
func golemOriginalByteAddress549E90(table unsafe.Pointer, i int32) unsafe.Pointer {
	return C.golem_byte_address_549e90(table, C.int32_t(i))
}
func golemOriginalBadName549E90() unsafe.Pointer { return C.golem_bad_name_549e90() }
func golemOriginalControl549E90(unit *server.Object) (out [4]uintptr, words [5]uint32) {
	var p [4]C.uintptr_t
	var w [5]C.uint32_t
	C.golem_control_549e90(asObjectC(unit), &p[0], &w[0])
	for i := range out {
		out[i] = uintptr(p[i])
	}
	for i := range words {
		words[i] = uint32(w[i])
	}
	return
}
func golemOriginalNew549E90(name unsafe.Pointer) *server.Object {
	return asObjectS(C.golem_new_549e90((*C.char)(name)))
}

var golemOriginalRuntimeOverride549E90 *server.GolemDeathRuntime549E90
var golemOriginalStateOverride549E90 *server.GolemDeathState549E90

func golemOriginalMassBytes549E90() []byte {
	return unsafe.Slice((*byte)(C.golem_mass_bytes_549e90()), 48)
}
