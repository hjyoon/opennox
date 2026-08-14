package legacy

/*
#include "audio_event_collide_4eaad0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export sub_4EAAD0
func sub_4EAAD0(source, target *C.nox_object_t, collision *C.float) {
	GetServer().S().AudioEventCollide4EAAD0(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
	)
}
