package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
#include "GAME1_1.h"
int nox_xxx_mobActionGet_50A020(nox_object_t* unit);
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t started; } idle_stack_5469b0;
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
 idle_stack_5469b0 stack138[24];
 uint32_t stamina282;
 void* seen283[16];
 void* enemy299;
 uint32_t fields300[4];
 void* preferred304;
 uint32_t field305;
 uint32_t scripts306[20];
 float aggression326;
} idle_layout_5469b0;
_Static_assert(_Generic(&nox_xxx_mobAction_5469B0,void (*)(nox_object_t*):1,default:0),"native idle void ABI");
_Static_assert(_Generic(&nox_xxx_monsterGetSoundSet_424300,void* (*)(nox_object_t*):1,default:0),"native sound-set pointer ABI");
static uintptr_t idle_original_public_5469b0(int root,nox_object_t* unit) {
 if(root)return (uintptr_t)nox_xxx_monsterGetSoundSet_424300(unit);
 nox_xxx_mobAction_5469B0(unit);return 0;
}
static uint32_t idle_original_action_5469b0(nox_object_t* unit) { return (uint32_t)nox_xxx_mobActionGet_50A020(unit); }
static uintptr_t idle_original_layout_5469b0(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,data_update);
 case 2:return offsetof(idle_layout_5469b0,sound122);
 case 3:return offsetof(idle_layout_5469b0,fields123)+5*4;
 case 4:return offsetof(idle_layout_5469b0,distance131);
 case 5:return offsetof(idle_layout_5469b0,timer132);
 case 6:return offsetof(idle_layout_5469b0,index136);
 case 7:return offsetof(idle_layout_5469b0,stack138);
 case 8:return offsetof(idle_layout_5469b0,enemy299);
 case 9:return offsetof(idle_layout_5469b0,aggression326);
 default:return sizeof(idle_stack_5469b0);
 }
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

type monsterIdleOriginalRuntime5469B0 struct {
	Frame, TickRate func() uint32
	Random          func(int, int) int
	Audio           func(uint32, *server.Object, uint32, uint32)
}

var monsterIdleRuntimeOverride5469B0 *monsterIdleOriginalRuntime5469B0

func idleOriginalC5469B0(root string, unit *server.Object) uintptr {
	kind := 0
	if root == "SoundSet" {
		kind = 1
	} else if root != "Idle" {
		panic("unknown idle original root")
	}
	return uintptr(C.idle_original_public_5469b0(C.int(kind), (*C.nox_object_t)(unsafe.Pointer(unit))))
}
func idleOriginalSoundSetGo5469B0(unit *server.Object) uintptr {
	return uintptr(nox_xxx_monsterGetSoundSet_424300((*C.nox_object_t)(unsafe.Pointer(unit))))
}
func idleOriginalActionC5469B0(unit *server.Object) uint32 {
	return uint32(C.idle_original_action_5469b0((*C.nox_object_t)(unsafe.Pointer(unit))))
}
func idleOriginalLayout5469B0() [11]uintptr {
	var out [11]uintptr
	for i := range out {
		out[i] = uintptr(C.idle_original_layout_5469b0(C.int(i)))
	}
	return out
}
