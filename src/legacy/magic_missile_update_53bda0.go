package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

var magicMissileUpdateRuntime53BDA0 = func() server.MagicMissileUpdateRuntime53BDA0 {
	world := GetServer().S()
	return server.MagicMissileUpdateRuntime53BDA0{
		Frame: world.Frame,
		FPS:   world.TickRate,
		SearchTarget: func(missile, owner *server.Object) *server.Object {
			return world.Nox_xxx_spellFlySearchTarget(nil, missile, things.SpellOffensive, 600, 0, owner)
		},
		Collide: func(missile *server.Object) { missile.CallCollide(0, 0) },
		CollideResult: func(missile *server.Object) int32 {
			return int32(ccall.CallIntPtr3(missile.Collide, missile.CObj(), nil, nil))
		},
		DirectionFloat: func(index int32) float32 {
			return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4)))
		},
	}
}

func magicMissileUpdateCall53BDA0(missile *server.Object) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	runtime := magicMissileUpdateRuntime53BDA0()
	// Registered native Go collision handlers keep their existing dispatcher.
	// The raw C entry retains the actual indirect C callback and its result.
	runtime.CollideResult = nil
	server.MagicMissileUpdate53BDA0(missile, runtime)
}

//export nox_magic_missile_update_native_53BDA0
func nox_magic_missile_update_native_53BDA0(missile *nox_object_t) C.int {
	return C.int(server.MagicMissileUpdateRaw53BDA0(asObjectS(missile), magicMissileUpdateRuntime53BDA0()))
}
