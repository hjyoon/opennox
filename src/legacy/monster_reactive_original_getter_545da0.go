package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME5.h"
typedef struct {
 uint32_t prefix[75];
 nox_waypoint_t* waypoints[16];
 uint32_t index,field92,field93,direction94;
 float2 pos95;
 uint32_t heard97,field98;
 float x99,y99;
 uint32_t heard_frame101;
} reactive_layout_545da0;
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t field5; } reactive_stack_545da0;
_Static_assert(_Generic(&sub_545DA0,int (*)(nox_object_t*):1,default:0),"native recent-sound ABI");
_Static_assert(_Generic(&nox_xxx_monsterLookAtDamager_5466B0,int (*)(nox_object_t*):1,default:0),"native look-damager ABI");
static int reactive_public_545da0(int root,nox_object_t* unit) {
 if(root==0)return sub_545DA0(unit);
 return nox_xxx_monsterLookAtDamager_5466B0(unit);
}
static uintptr_t reactive_layout_value_545da0(int field) {
 switch(field) {
 case 0:return offsetof(reactive_layout_545da0,heard97);
 case 1:return offsetof(reactive_layout_545da0,x99);
 case 2:return offsetof(reactive_layout_545da0,heard_frame101);
 case 3:return offsetof(nox_object_t,field_132);
 case 4:return offsetof(nox_object_t,data_update);
 case 5:return sizeof(reactive_stack_545da0);
 case 6:return offsetof(reactive_stack_545da0,args);
 default:return offsetof(reactive_stack_545da0,field5);
 }
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var monsterReactiveRuntimeOverride545DA0 *server.MonsterReactiveRuntime545DA0

func monsterReactiveOriginalC545DA0(root string, unit *server.Object) int {
	index := 0
	if root == "Look" {
		index = 1
	} else if root != "Sound" {
		panic("unknown reactive root")
	}
	return int(C.reactive_public_545da0(C.int(index), (*C.nox_object_t)(unsafe.Pointer(unit))))
}
func monsterReactiveOriginalLayout545DA0() [8]uintptr {
	var out [8]uintptr
	for i := range out {
		out[i] = uintptr(C.reactive_layout_value_545da0(C.int(i)))
	}
	return out
}
