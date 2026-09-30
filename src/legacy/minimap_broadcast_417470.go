package legacy

/*
#include "GAME1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func minimapMarkBroadcastCall4174B0(obj *server.Object, flags uint32) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_netMarkMinimapForAll_4174B0(asObjectC(obj), C.int(flags)))
}

func minimapUnmarkBroadcastCall417470(obj *server.Object, flags uint32) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_netUnmarkMinimapSpec_417470(asObjectC(obj), C.int(flags)))
}
