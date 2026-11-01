package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&sub_519710, int (*)(void*):1, default:0), "original tracked refresh ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func trackedRefreshOriginalPublicC519710(update *server.PlayerUpdateData) int32 {
	return int32(C.sub_519710(unsafe.Pointer(update)))
}
