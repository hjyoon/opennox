package legacy

/*
#include "GAME4_3.h"
#include <stddef.h>
typedef void (*monster_direction_callback_C)(nox_object_t*, float2*);
_Static_assert(_Generic(&nox_xxx_mobCalcDir_533CC0, monster_direction_callback_C:1, default:0), "native monster direction ABI");
static void monster_direction_layout_C(uintptr_t* v) {
 v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,x);v[2]=offsetof(nox_object_t,y);
 v[3]=offsetof(nox_object_t,vel_x);v[4]=offsetof(nox_object_t,vel_y);
 v[5]=offsetof(nox_object_t,direction1);v[6]=offsetof(nox_object_t,direction2);
 v[7]=sizeof(float2);v[8]=offsetof(float2,field_4);
}
static void monster_direction_call_C(int callback,nox_object_t* unit,float2* position) {
 if(callback){monster_direction_callback_C f=nox_xxx_mobCalcDir_533CC0;f(unit,position);}
 else nox_xxx_mobCalcDir_533CC0(unit,position);
}
*/
import "C"
import "unsafe"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func monsterDirectionLayoutC() (v [9]uintptr) {
	C.monster_direction_layout_C((*C.uintptr_t)(unsafe.Pointer(&v[0])))
	return v
}

func monsterDirectionActualCall(route int, unit *server.Object, position *types.Pointf) {
	if route == 2 {
		monsterCastDefaultRuntime541300().Direction(unit, position)
		return
	}
	C.monster_direction_call_C(C.int(route), asObjectC(unit), (*C.float2)(unsafe.Pointer(position)))
}
