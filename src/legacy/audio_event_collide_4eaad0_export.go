package legacy

/*
#include "audio_event_collide_4eaad0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var audioEventCollideCall4EAAD0 = func(source, target *server.Object, collision unsafe.Pointer) {
	GetServer().S().AudioEventCollide4EAAD0(source, target, (*types.Pointf)(collision))
}

//export sub_4EAAD0
func sub_4EAAD0(source, target *C.nox_object_t, collision *C.float) {
	audioEventCollideCall4EAAD0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	)
}
