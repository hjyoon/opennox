package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var trackedRefreshRuntimeNative519710 = func() server.TrackedRefreshRuntime519710 {
	s := GetServer().S()
	return server.TrackedRefreshRuntime519710{Count: sub_417270, Frame: s.Frame}
}

func TrackedObjectRefresh519710(update *server.PlayerUpdateData) int32 {
	return server.TrackedRefreshNative519710(update, trackedRefreshRuntimeNative519710())
}

//export sub_519710
func sub_519710(update unsafe.Pointer) C.int {
	return C.int(TrackedObjectRefresh519710((*server.PlayerUpdateData)(update)))
}
