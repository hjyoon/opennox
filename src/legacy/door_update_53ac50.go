package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_updateDoor_53AC50
func nox_xxx_updateDoor_53AC50(door *nox_object_t) C.char {
	return C.char(doorUpdateResult53AC50(asObjectS(door)))
}

func doorUpdateCall53AC50(door *server.Object) {
	doorUpdateResult53AC50(door)
}

func doorUpdateResult53AC50(door *server.Object) byte {
	world := GetServer().S()
	return world.DoorUpdate53AC50(door, server.DoorUpdateRuntime53AC50{
		AudioEvent: func(id uint32, obj *server.Object) {
			world.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		QueueDoor: game5EnqueueDoor,
		WakeDoor: func(obj *server.Object) byte {
			return byte(pendingCollisionEnqueue537610(obj))
		},
	})
}
