package legacy

/*

#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
#include "GAME1_1.h"
int nox_xxx_mobActionGet_50A020(nox_object_t* unit);
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } pop_stack_5471b0;
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
 pop_stack_5471b0 stack138[24];
 uint32_t stamina282;
 void* seen283[16];
 void* enemy299;
 uint32_t fields300[4];
 void* preferred304;
 uint32_t field305;
 uint32_t scripts306[20];
 float aggression326;
} pop_layout_5471b0;

static void pop_legacy_5471b0(void (*fn)(int),nox_object_t* unit) { fn((int)(uintptr_t)unit); }
static void pop_native_5471b0(void (*fn)(nox_object_t*),nox_object_t* unit) { fn(unit); }
static int pop_abi_5471b0(void) { return _Generic(&nox_xxx_monsterPopAttackActions_5471B0,void (*)(int):32,void (*)(nox_object_t*):64); }
static void pop_original_public_5471b0(nox_object_t* unit) {
 _Generic(&nox_xxx_monsterPopAttackActions_5471B0,void (*)(int):pop_legacy_5471b0,void (*)(nox_object_t*):pop_native_5471b0)(&nox_xxx_monsterPopAttackActions_5471B0,unit);
}
static uintptr_t pop_control_5471b0(nox_object_t* unit) { return (uintptr_t)nox_xxx_monsterGetSoundSet_424300(unit); }
static uint32_t pop_action_5471b0(nox_object_t* unit) { return nox_xxx_mobActionGet_50A020(unit); }
static uintptr_t pop_layout_offset_5471b0(int field) {
 switch(field) {
 case 0:return sizeof(nox_object_t);
 case 1:return offsetof(nox_object_t,data_update);
 case 2:return offsetof(pop_layout_5471b0,timer132);
 case 3:return offsetof(pop_layout_5471b0,index136);
 case 4:return offsetof(pop_layout_5471b0,stack138);
 case 5:return sizeof(pop_stack_5471b0);
 case 6:return offsetof(pop_stack_5471b0,args);
 default:return offsetof(pop_stack_5471b0,started);
 }
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var monsterPopAttacksOverride5471B0 func(*server.Object) int

func popOriginalC5471B0(unit *server.Object) {
	C.pop_original_public_5471b0((*C.nox_object_t)(unsafe.Pointer(unit)))
}
func popOriginalABI5471B0() int { return int(C.pop_abi_5471b0()) }
func popOriginalControl5471B0(unit *server.Object) uintptr {
	return uintptr(C.pop_control_5471b0((*C.nox_object_t)(unsafe.Pointer(unit))))
}
func popOriginalAction5471B0(unit *server.Object) uint32 {
	return uint32(C.pop_action_5471b0((*C.nox_object_t)(unsafe.Pointer(unit))))
}
func popOriginalLayout5471B0() [8]uintptr {
	var out [8]uintptr
	for i := range out {
		out[i] = uintptr(C.pop_layout_offset_5471b0(C.int(i)))
	}
	return out
}
