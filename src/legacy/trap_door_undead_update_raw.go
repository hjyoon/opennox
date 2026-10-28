package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Resolve the world only at reached service boundaries. A cancelled duration
// record does not read the clock; an enabled opening needs no collision data.
var trapDoorUpdateRawRuntime53DE80 = func() server.TrapDoorUpdateRawRuntime53DE80 {
	return server.TrapDoorUpdateRawRuntime53DE80{
		Frame:        func() uint32 { return GetServer().S().Frame() },
		FPS:          func() uint32 { return GetServer().S().TickRate() },
		SetOnOff:     (*server.Object).SetOnOff,
		UnsetXStatus: (*server.Object).UnsetXStatus, SetXStatus: (*server.Object).SetXStatus,
		AudioEvent: func(id uint32, o *server.Object) { GetServer().S().Audio.EventObj(sound.ID(id), o, 0, 0) },
	}
}

var undeadKillerUpdateRawRuntime53E190 = func() server.UndeadKillerUpdateRawRuntime53E190 {
	return server.UndeadKillerUpdateRawRuntime53E190{
		Frame:         func() uint32 { return GetServer().S().Frame() },
		DelayedDelete: func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

//export nox_trapdoor_update_native_53de80
func nox_trapdoor_update_native_53de80(source *nox_object_t) {
	server.TrapDoorUpdateRaw53DE80(asObjectS(source), trapDoorUpdateRawRuntime53DE80())
}

//export nox_undead_killer_update_native_53e190
func nox_undead_killer_update_native_53e190(source *nox_object_t) {
	server.UndeadKillerUpdateRaw53E190(asObjectS(source), undeadKillerUpdateRawRuntime53E190())
}
