package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME5.h"
#include "map_push_units_52e040.h"

static int monster_death_original_old_549d80(int (*fn)(int), nox_object_t* unit) {
 return fn((int)(uintptr_t)unit);
}
static int monster_death_original_native_549d80(int (*fn)(nox_object_t*), nox_object_t* unit) {
 return fn(unit);
}
#define MONSTER_DEATH_ORIGINAL_CALL_549D80(fn, unit) \
 _Generic(&(fn), int (*)(int): monster_death_original_old_549d80, int (*)(nox_object_t*): monster_death_original_native_549d80)(&(fn), (unit))
#define MONSTER_DEATH_ORIGINAL_ABI_549D80(fn) \
 _Generic(&(fn), int (*)(int): 32, int (*)(nox_object_t*): 64)
static int monster_death_original_public_549d80(uint32_t entry, nox_object_t* unit) {
 switch (entry) {
 case 0x549d80: return MONSTER_DEATH_ORIGINAL_CALL_549D80(sub_549D80, unit);
 case 0x549e00: return MONSTER_DEATH_ORIGINAL_CALL_549D80(sub_549E00, unit);
 case 0x549e70: return MONSTER_DEATH_ORIGINAL_CALL_549D80(sub_549E70, unit);
 case 0x54a250: return MONSTER_DEATH_ORIGINAL_CALL_549D80(sub_54A250, unit);
 default: return -1;
 }
}
static int monster_death_original_abi_549d80(uint32_t entry) {
 switch (entry) {
 case 0x549d80: return MONSTER_DEATH_ORIGINAL_ABI_549D80(sub_549D80);
 case 0x549e00: return MONSTER_DEATH_ORIGINAL_ABI_549D80(sub_549E00);
 case 0x549e70: return MONSTER_DEATH_ORIGINAL_ABI_549D80(sub_549E70);
 case 0x54a250: return MONSTER_DEATH_ORIGINAL_ABI_549D80(sub_54A250);
 default: return -1;
 }
}
static uintptr_t monster_death_original_layout_549d80(int index) {
 switch(index) {
 case 0: return offsetof(nox_object_t,x);
 case 1: return offsetof(nox_object_t,y);
 case 2: return offsetof(nox_object_t,script_pickup_func) + sizeof(((nox_object_t*)0)->script_pickup_func);
 default: return sizeof(nox_object_t);
 }
}
static nox_object_t* monster_death_original_bad_object_549d80(int role) {
 return (nox_object_t*)(uintptr_t)(role == -1 ? 0 : 0x41000000);
}
static void monster_death_original_controls_549d80(nox_object_t* unit, uintptr_t* out) {
 out[0] = (uintptr_t)unit;
 out[1] = (uintptr_t)&unit->x;
 out[2] = (uintptr_t)&unit->y;
}
static void monster_death_original_push_control_549d80(nox_object_t* unit) {
 nox_xxx_mapPushUnitsAround_52E040(&unit->x,96.0f,10.0f,100.0f,unit,0,0);
}
static uint32_t monster_death_original_point_control_549d80(nox_object_t* unit) {
 return (uint32_t)nox_xxx_netSendPointFx_522FF0((char)0x81,(float2*)&unit->x);
}
static uint32_t monster_death_original_spark_control_549d80(nox_object_t* unit) {
 return (uint32_t)nox_xxx_netSparkExplosionFx_5231B0(&unit->x,(char)0x80);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func monsterDeathOriginalC549D80(entry uint32, unit *server.Object) int32 {
	return int32(C.monster_death_original_public_549d80(C.uint32_t(entry), (*C.nox_object_t)(unsafe.Pointer(unit))))
}

func monsterDeathOriginalABI549D80(entry uint32) int {
	return int(C.monster_death_original_abi_549d80(C.uint32_t(entry)))
}

func monsterDeathOriginalLayout549D80() (out [4]uintptr) {
	for i := range out {
		out[i] = uintptr(C.monster_death_original_layout_549d80(C.int(i)))
	}
	return
}

func monsterDeathOriginalBadObject549D80(role int) *server.Object {
	return (*server.Object)(unsafe.Pointer(C.monster_death_original_bad_object_549d80(C.int(role))))
}

func monsterDeathOriginalControls549D80(unit *server.Object) (out [3]uintptr) {
	var raw [3]C.uintptr_t
	C.monster_death_original_controls_549d80((*C.nox_object_t)(unsafe.Pointer(unit)), &raw[0])
	for i := range out {
		out[i] = uintptr(raw[i])
	}
	return
}

func monsterDeathOriginalPointControl549D80(unit *server.Object) uint32 {
	return uint32(C.monster_death_original_point_control_549d80((*C.nox_object_t)(unsafe.Pointer(unit))))
}

func monsterDeathOriginalSparkControl549D80(unit *server.Object) uint32 {
	return uint32(C.monster_death_original_spark_control_549d80((*C.nox_object_t)(unsafe.Pointer(unit))))
}

var monsterDeathOriginalRuntimeOverride549D80 *server.MonsterDeathExplosionRuntime549D80

func monsterDeathOriginalPushControl549D80(unit *server.Object) {
	C.monster_death_original_push_control_549d80((*C.nox_object_t)(unsafe.Pointer(unit)))
}
