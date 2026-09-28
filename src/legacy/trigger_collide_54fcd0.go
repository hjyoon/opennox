package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func triggerCollideCompatibilityAllowed54FCD0(mapName, callbackName, candidateID string) bool {
	mapName = strings.TrimSpace(strings.ToLower(mapName))
	if i := strings.LastIndexAny(mapName, `/\\`); i >= 0 {
		mapName = mapName[i+1:]
	}
	mapName = strings.TrimSuffix(mapName, ".map")
	if mapName != "con02a" || !strings.EqualFold(callbackName, "NecroInPosition") {
		return true
	}

	// Con02A places the NecroInPosition trigger on the route used by several
	// other actors before and during the setpiece. The callback advances the
	// scene unconditionally, so a small native-width routing/timing difference
	// can let Tanya2 or Julie2 consume it before the Necromancer arrives.
	// This arrival callback belongs exclusively to the Necromancer; the four
	// actor-count triggers used by the later spider summon are separate objects.
	if i := strings.LastIndexByte(candidateID, ':'); i >= 0 {
		candidateID = candidateID[i+1:]
	}
	return strings.EqualFold(candidateID, "Necromancer")
}

func triggerActivateCallbackName54FCD0(srv *server.Server, trigger *server.Object) string {
	if srv == nil || trigger == nil || trigger.UpdateData == nil {
		return ""
	}
	index := int(trigger.UpdateDataTrigger().ScriptActivate.Func)
	funcs := srv.NoxScriptVM.Funcs()
	if index < 0 || index >= len(funcs) {
		return ""
	}
	return funcs[index].Name()
}

var triggerCollideCall54FCD0 = func(trigger, candidate *server.Object, _ unsafe.Pointer) {
	srv := GetServer()
	if candidate != nil && !triggerCollideCompatibilityAllowed54FCD0(
		Nox_xxx_mapGetMapName_409B40(),
		triggerActivateCallbackName54FCD0(srv.S(), trigger),
		candidate.ID(),
	) {
		return
	}
	srv.S().TriggerCollide54FCD0(trigger, candidate, server.TriggerCollideRuntime54FCD0{
		Mass: objectMassC,
		ScriptAllowed: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) bool {
			result := srv.NoxScriptC().ScriptCallback(block, caller, trigger, event)
			return result != nil && *(*uint32)(unsafe.Pointer(result)) != 0
		},
	})
}

//export nox_xxx_collideTrigger_54FCD0_go
func nox_xxx_collideTrigger_54FCD0_go(trigger, candidate *nox_object_t) {
	triggerCollideCall54FCD0(asObjectS(trigger), asObjectS(candidate), nil)
}
