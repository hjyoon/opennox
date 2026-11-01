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
 uint32_t marker97,field98;
 float x99,y99;
 uint32_t frame101;
} interesting_layout_5466f0;
typedef struct { uint32_t action; uintptr_t args[4]; uint32_t field5; } interesting_stack_5466f0;
_Static_assert(_Generic(&sub_5466F0,int (*)(nox_object_t*):1,default:0),"native interesting-sound public ABI");
static int interesting_public_5466f0(nox_object_t* unit) { return sub_5466F0(unit); }
static uintptr_t interesting_layout_value_5466f0(int field) {
 switch(field) {
 case 0:return offsetof(nox_object_t,obj_class);
 case 1:return offsetof(nox_object_t,obj_subclass);
 case 2:return offsetof(nox_object_t,x);
 case 3:return offsetof(nox_object_t,data_update);
 case 4:return offsetof(interesting_layout_5466f0,marker97);
 case 5:return offsetof(interesting_layout_5466f0,x99);
 case 6:return offsetof(interesting_layout_5466f0,frame101);
 case 7:return sizeof(interesting_stack_5466f0);
 case 8:return offsetof(interesting_stack_5466f0,args);
 default:return offsetof(interesting_stack_5466f0,field5);
 }
}
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

// A comparison-only override; ordinary production never sets it.
type monsterInterestingSoundOriginalRuntime5466F0 struct {
	Reactive server.MonsterReactiveRuntime545DA0
	Tile     func(*types.Pointf) int
	Path     func(*server.Object, *types.Pointf) bool
	Trace    func(types.Pointf, types.Pointf) bool
}

var monsterInterestingSoundRuntimeOverride5466F0 *monsterInterestingSoundOriginalRuntime5466F0

func monsterInterestingOriginalC5466F0(unit *server.Object) int {
	return int(C.interesting_public_5466f0((*C.nox_object_t)(unsafe.Pointer(unit))))
}

func monsterInterestingOriginalLayout5466F0() [10]uintptr {
	var out [10]uintptr
	for i := range out {
		out[i] = uintptr(C.interesting_layout_value_5466f0(C.int(i)))
	}
	return out
}
