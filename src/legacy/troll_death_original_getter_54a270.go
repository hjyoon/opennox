package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5.h"

static int troll_death_old_54a270(int (*fn)(int), nox_object_t* unit) { return fn((int)(uintptr_t)unit); }
static int troll_death_native_54a270(int (*fn)(nox_object_t*), nox_object_t* unit) { return fn(unit); }
static int troll_death_public_54a270(nox_object_t* unit) {
 return _Generic(&nox_xxx_monsterDeadTroll_54A270,
  int (*)(int): troll_death_old_54a270,
  int (*)(nox_object_t*): troll_death_native_54a270)(&nox_xxx_monsterDeadTroll_54A270,unit);
}
static int troll_death_abi_54a270(void) {
 return _Generic(&nox_xxx_monsterDeadTroll_54A270,int (*)(int):32,int (*)(nox_object_t*):64);
}
static uintptr_t troll_death_layout_54a270(int index) {
 switch(index) {
 case 0:return offsetof(nox_object_t,x);
 case 1:return offsetof(nox_object_t,y);
 case 2:return offsetof(nox_object_t,data_update);
 default:return sizeof(nox_object_t);
 }
}
static void* troll_death_bad_54a270(int role) { return (void*)(uintptr_t)(role == -1 ? 0 : 0x41000000); }
static void troll_death_control_54a270(nox_object_t* unit,uintptr_t* out,uint32_t* bits) {
 out[0]=(uintptr_t)unit;out[1]=(uintptr_t)unit->data_update;
 out[2]=(uintptr_t)&unit->x;out[3]=(uintptr_t)&unit->y;
 memcpy(bits,&unit->x,8);
}
static uint32_t troll_death_type_54a270(void) { return (uint32_t)nox_xxx_getNameId_4E3AA0("SmallToxicCloud"); }
static nox_object_t* troll_death_new_54a270(uint32_t id) { return nox_xxx_newObjectWithTypeInd_4E3450((int32_t)id); }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func trollDeathOriginalC54A270(unit *server.Object) int32 {
	return int32(C.troll_death_public_54a270((*C.nox_object_t)(unsafe.Pointer(unit))))
}

func trollDeathOriginalABI54A270() int { return int(C.troll_death_abi_54a270()) }

func trollDeathOriginalLayout54A270() (out [4]uintptr) {
	for i := range out {
		out[i] = uintptr(C.troll_death_layout_54a270(C.int(i)))
	}
	return
}

func trollDeathOriginalBad54A270(role int) unsafe.Pointer {
	return C.troll_death_bad_54a270(C.int(role))
}

func trollDeathOriginalControl54A270(unit *server.Object) (out [4]uintptr, bits [2]uint32) {
	var cOut [4]C.uintptr_t
	var cBits [2]C.uint32_t
	C.troll_death_control_54a270((*C.nox_object_t)(unsafe.Pointer(unit)), &cOut[0], &cBits[0])
	for i := range out {
		out[i] = uintptr(cOut[i])
	}
	for i := range bits {
		bits[i] = uint32(cBits[i])
	}
	return
}

func trollDeathOriginalType54A270() uint32 { return uint32(C.troll_death_type_54a270()) }
func trollDeathOriginalNew54A270(id uint32) *server.Object {
	return (*server.Object)(unsafe.Pointer(C.troll_death_new_54a270(C.uint32_t(id))))
}

var trollDeathOriginalRuntimeOverride54A270 *server.TrollDeathRuntime54A270
var trollDeathOriginalStateOverride54A270 *server.TrollDeathState54A270
