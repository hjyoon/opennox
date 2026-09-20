package legacy

/*
#include "crown_collide_4ebb50.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func crownCollideRuntime4EBB50(s *server.Server) server.CrownCollideRuntime4EBB50 {
	return server.CrownCollideRuntime4EBB50{
		Pickup: func(who, crown *server.Object, flag1, flag2 int32) uint32 {
			return crownPickupCall4F3400(s, who, crown, flag1, flag2)
		},
	}
}

var crownCollideCall4EBB50 = func(crown, target *server.Object, collision unsafe.Pointer) uintptr {
	s := GetServer().S()
	return s.CrownCollide4EBB50(
		crown,
		target,
		(*types.Pointf)(collision),
		crownCollideRuntime4EBB50(s),
	)
}

//export sub_4EBB50
func sub_4EBB50(
	crown, target *C.nox_object_t,
	collision *C.float,
) C.uintptr_t {
	return C.uintptr_t(crownCollideCall4EBB50(
		asObjectS((*nox_object_t)(crown)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	))
}
