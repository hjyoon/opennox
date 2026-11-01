package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func monsterIdleSoundRuntime5469B0() monsterIdleOriginalRuntime5469B0 {
	if override := monsterIdleRuntimeOverride5469B0; override != nil {
		return *override
	}
	return monsterIdleOriginalRuntime5469B0{
		Frame:    func() uint32 { return GetServer().S().Frame() },
		TickRate: func() uint32 { return GetServer().S().TickRate() },
		Random:   func(min, max int) int { return GetServer().S().Rand.Logic.IntClamp(min, max) },
		Audio: func(id uint32, unit *server.Object, a, b uint32) {
			GetServer().S().Audio.EventObj(sound.ID(id), unit, int(a), b)
		},
	}
}

//export nox_xxx_mobAction_5469B0
func nox_xxx_mobAction_5469B0(unit *C.nox_object_t) {
	runtime := monsterIdleSoundRuntime5469B0()
	(*server.Object)(unsafe.Pointer(unit)).MonsterIdleSoundNative5469B0(runtime.Frame, runtime.TickRate, runtime.Random, runtime.Audio)
}
