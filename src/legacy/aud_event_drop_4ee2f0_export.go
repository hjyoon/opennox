package legacy

/*
#include "aud_event_drop_4ee2f0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_objectDropAudEvent_4EE2F0
func nox_objectDropAudEvent_4EE2F0(
	owner, item *C.nox_object_t,
	point *C.float2,
) C.int {
	return C.int(audEventDropCall4EE2F0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		(*types.Pointf)(unsafe.Pointer(point)),
	))
}
