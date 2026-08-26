package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideTrigger_54FCD0_go
func nox_xxx_collideTrigger_54FCD0_go(trigger, candidate *nox_object_t) {
	srv := GetServer()
	srv.S().TriggerCollide54FCD0(asObjectS(trigger), asObjectS(candidate), server.TriggerCollideRuntime54FCD0{
		Mass: objectMassC,
		ScriptAllowed: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) bool {
			result := srv.NoxScriptC().ScriptCallback(block, caller, trigger, event)
			return result != nil && *(*uint32)(unsafe.Pointer(result)) != 0
		},
	})
}
