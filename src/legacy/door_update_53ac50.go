package legacy

/*
#include "GAME4_3.h"
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func doorUpdateCall53AC50(door *server.Object) {
	world := GetServer().S()
	world.DoorUpdate53AC50(door, server.DoorUpdateRuntime53AC50{
		AudioEvent: func(id uint32, obj *server.Object) {
			world.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		QueueDoor: func(update *server.DoorUpdateData) {
			C.sub_548830(C.uintptr_t(uintptr(unsafe.Pointer(update))))
		},
		WakeDoor: Nox_xxx_unitHasCollideOrUpdateFn_537610,
	})
}
