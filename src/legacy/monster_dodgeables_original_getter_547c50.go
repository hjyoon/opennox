package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
#include "GAME1_1.h"
int nox_xxx_mobActionGet_50A020(nox_object_t* unit);
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } dodge_stack_547c50;
typedef struct {
 uint32_t prefix[75];
 void* waypoints[16];
 uint32_t fields91[12];
 uint16_t health103[32];
 void* animation119;
 uint32_t flags120;
 void* monster121;
 void* sound122;
 uint32_t fields123[8];
 uint32_t distance131,timer132,fields133[3];
 int8_t index136; uint8_t pad136; uint16_t pad136b;
 uint32_t frame137;
 dodge_stack_547c50 stack138[24];
} dodge_layout_547c50;
static int dodge_legacy_547c50(int (*fn)(int),nox_object_t* unit) { return fn((int)(uintptr_t)unit); }
static int dodge_native_547c50(int (*fn)(nox_object_t*),nox_object_t* unit) { return fn(unit); }
static int dodge_abi_547c50(void) { return _Generic(&nox_xxx_monsterCheckDodgeables_547C50,int (*)(int):32,int (*)(nox_object_t*):64); }
static int dodge_original_public_547c50(nox_object_t* unit) {
 return _Generic(&nox_xxx_monsterCheckDodgeables_547C50,int (*)(int):dodge_legacy_547c50,int (*)(nox_object_t*):dodge_native_547c50)(&nox_xxx_monsterCheckDodgeables_547C50,unit);
}
static uintptr_t dodge_control_547c50(nox_object_t* unit) { return (uintptr_t)nox_xxx_monsterGetSoundSet_424300(unit); }
static uint32_t dodge_action_547c50(nox_object_t* unit) { return nox_xxx_mobActionGet_50A020(unit); }
static uintptr_t dodge_layout_offset_547c50(int field) {
 switch(field) {
 case 0:return sizeof(nox_object_t);
 case 1:return offsetof(nox_object_t,x);
 case 2:return offsetof(nox_object_t,y);
 case 3:return offsetof(nox_object_t,direction1);
 case 4:return offsetof(nox_object_t,shape)+offsetof(nox_shape,circle_r);
 case 5:return offsetof(nox_object_t,speed_cur);
 case 6:return offsetof(nox_object_t,data_update);
 case 7:return offsetof(dodge_layout_547c50,timer132);
 case 8:return offsetof(dodge_layout_547c50,index136);
 case 9:return offsetof(dodge_layout_547c50,stack138);
 case 10:return sizeof(dodge_stack_547c50);
 case 11:return offsetof(dodge_stack_547c50,args);
 default:return offsetof(dodge_stack_547c50,started);
 }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

type monsterDodgeOriginalRuntime547C50 struct {
	RandomFloat func(float32, float32) float64
	RandomInt   func(int, int) int
	Trace       func(*[4]float32, uint32) int
	Obstacles   func(*server.Object, *[4]float32) int
	Tile        func(*types.Pointf) int
	Pop         func(*server.Object) int
	Push        func(*server.Object, ai.ActionType) *server.AIStackItem
	Frame, FPS  func() uint32
}

var monsterDodgeRuntimeOverride547C50 *monsterDodgeOriginalRuntime547C50

func dodgeOriginalC547C50(unit *server.Object) int {
	return int(C.dodge_original_public_547c50((*C.nox_object_t)(unsafe.Pointer(unit))))
}
func dodgeOriginalABI547C50() int { return int(C.dodge_abi_547c50()) }
func dodgeOriginalControls547C50(unit *server.Object) (uintptr, uint32) {
	u := (*C.nox_object_t)(unsafe.Pointer(unit))
	return uintptr(C.dodge_control_547c50(u)), uint32(C.dodge_action_547c50(u))
}
func dodgeOriginalLayout547C50() [13]uintptr {
	var out [13]uintptr
	for i := range out {
		out[i] = uintptr(C.dodge_layout_offset_547c50(C.int(i)))
	}
	return out
}
