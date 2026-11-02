package legacy

/*
#include "defs.h"
#include "GAME5.h"
_Static_assert(_Generic(&sub_5456D0,void (*)(nox_object_t*):1,default:0), "native raw retreat-to-master void ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

//export sub_5456D0
func sub_5456D0(unit *C.nox_object_t) {
	obj := (*server.Object)(unsafe.Pointer(unit))
	push := func(action ai.ActionType) *server.AIStackItem { return obj.MonsterPushAction(action) }
	pop := func() int { return obj.MonsterPopAction() }
	if hooks := retreatMasterRuntimeOverride5456D0; hooks != nil {
		push = func(action ai.ActionType) *server.AIStackItem { return hooks.Push(obj, action) }
		pop = func() int { return hooks.Pop(obj) }
	}
	obj.MonsterRetreatToMasterNative5456D0(push, pop)
}
