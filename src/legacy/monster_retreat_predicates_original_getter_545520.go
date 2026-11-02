package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } retreat_predicate_stack_545520;
typedef struct {
 uint32_t prefix[75]; void* waypoints[16]; uint32_t fields91[12];
 uint16_t health103[32]; void* animation119; uint32_t flags120;
 void* monster121; void* sound122; uint32_t fields123[13];
 int8_t index136; uint8_t pad136; uint16_t pad136b; uint32_t frame137;
 retreat_predicate_stack_545520 stack138[24];
 uint8_t stamina; uint8_t seen_count; uint16_t pad282; void* seen[16];
 void* enemy; uint32_t fields300[4]; void* preferred_enemy; uint32_t field305;
 uint32_t scripts306[20]; uint32_t fields326[10]; float resume;
 uint32_t fields337[23]; uint32_t status;
} retreat_predicate_update_545520;
typedef struct { uint16_t cur, field1, max; } retreat_predicate_health_545520;
static int retreat_predicate_legacy_545520(int (*fn)(int), nox_object_t* unit) { return fn((int)(uintptr_t)unit); }
static int retreat_predicate_native_545520(int (*fn)(nox_object_t*), nox_object_t* unit) { return fn(unit); }
#define RETREAT_PREDICATE_ABI(fn) _Generic(&(fn), int (*)(int):32, int (*)(nox_object_t*):64)
#define RETREAT_PREDICATE_CALL(fn,unit) _Generic(&(fn), int (*)(int):retreat_predicate_legacy_545520, int (*)(nox_object_t*):retreat_predicate_native_545520)(&(fn),unit)
static int retreat_original_abi_545520(int entry) {
 switch(entry) {
 case 0:return RETREAT_PREDICATE_ABI(nox_xxx_monsterCanResumeAttack_545520);
 case 1:return RETREAT_PREDICATE_ABI(sub_545580);
 default:return RETREAT_PREDICATE_ABI(nox_xxx_monsterCanCast2_5455B0);
 }
}
static int retreat_original_public_545520(int entry,nox_object_t* unit) {
 switch(entry) {
 case 0:return RETREAT_PREDICATE_CALL(nox_xxx_monsterCanResumeAttack_545520,unit);
 case 1:return RETREAT_PREDICATE_CALL(sub_545580,unit);
 default:return RETREAT_PREDICATE_CALL(nox_xxx_monsterCanCast2_5455B0,unit);
 }
}
static uintptr_t retreat_original_layout_545520(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,health_data);
 case 1:return offsetof(nox_object_t,data_update);
 case 2:return offsetof(nox_object_t,buffs);
 case 3:return offsetof(retreat_predicate_health_545520,max);
 case 4:return offsetof(retreat_predicate_update_545520,resume);
 default:return offsetof(retreat_predicate_update_545520,status);
 }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func retreatPredicatesOriginalC545520(entry int, unit *server.Object) int {
	return int(C.retreat_original_public_545520(C.int(entry), (*C.nox_object_t)(unsafe.Pointer(unit))))
}
func retreatPredicatesOriginalABI545520(entry int) int {
	return int(C.retreat_original_abi_545520(C.int(entry)))
}
func retreatPredicatesOriginalLayout545520() (out [6]uintptr) {
	for i := range out {
		out[i] = uintptr(C.retreat_original_layout_545520(C.int(i)))
	}
	return out
}
