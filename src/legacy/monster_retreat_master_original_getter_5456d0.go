package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } retreat_master_stack_5456d0;
typedef struct {
 uint32_t prefix[75]; void* waypoints[16]; uint32_t fields91[12];
 uint16_t health103[32]; void* animation119; uint32_t flags120;
 void* monster121; void* sound122; uint32_t fields123[13];
 int8_t index136; uint8_t pad136; uint16_t pad136b; uint32_t frame137;
 retreat_master_stack_5456d0 stack138[24];
 uint8_t stamina; uint8_t seen_count; uint16_t pad282; void* seen[16];
 void* enemy; uint32_t fields300[4]; void* preferred_enemy; uint32_t field305;
 uint32_t scripts306[20]; uint32_t fields326[3]; float range;
 uint32_t fields330[6]; float resume; uint32_t fields337[23]; uint32_t status;
} retreat_master_update_5456d0;
static void retreat_master_legacy_5456d0(void (*fn)(int),nox_object_t* unit) { fn((int)(uintptr_t)unit); }
static void retreat_master_native_5456d0(void (*fn)(nox_object_t*),nox_object_t* unit) { fn(unit); }
static int retreat_master_abi_5456d0(void) { return _Generic(&sub_5456D0,void (*)(int):32,void (*)(nox_object_t*):64); }
static void retreat_master_original_public_5456d0(nox_object_t* unit) {
 _Generic(&sub_5456D0,void (*)(int):retreat_master_legacy_5456d0,void (*)(nox_object_t*):retreat_master_native_5456d0)(&sub_5456D0,unit);
}
static uintptr_t retreat_master_layout_5456d0(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,owner);
 case 1:return offsetof(nox_object_t,data_update);
 case 2:return offsetof(nox_object_t,x);
 case 3:return offsetof(nox_object_t,health_data);
 case 4:return sizeof(nox_object_t);
 case 5:return offsetof(retreat_master_update_5456d0,range);
 case 6:return offsetof(retreat_master_update_5456d0,resume);
 case 7:return offsetof(retreat_master_update_5456d0,status);
 case 8:return sizeof(retreat_master_stack_5456d0);
 case 9:return offsetof(retreat_master_stack_5456d0,args);
 default:return offsetof(retreat_master_stack_5456d0,started);
 }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

type retreatMasterOriginalRuntime5456D0 struct {
	Push func(*server.Object, ai.ActionType) *server.AIStackItem
	Pop  func(*server.Object) int
}

var retreatMasterRuntimeOverride5456D0 *retreatMasterOriginalRuntime5456D0

func retreatMasterOriginalC5456D0(unit *server.Object) {
	C.retreat_master_original_public_5456d0((*C.nox_object_t)(unsafe.Pointer(unit)))
}
func retreatMasterOriginalABI5456D0() int { return int(C.retreat_master_abi_5456d0()) }
func retreatMasterOriginalLayout5456D0() (out [11]uintptr) {
	for i := range out {
		out[i] = uintptr(C.retreat_master_layout_5456d0(C.int(i)))
	}
	return out
}
