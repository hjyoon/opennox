package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&nox_xxx_netSpriteUpdate_518AE0, int (*)(nox_object_t*,int,uint32_t*):1,default:0), "native recipient pointer/int/opaque object pointer sprite ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func spriteUpdatePublicC518AE0(recipient *server.Object, player int32, unit *server.Object) int32 {
	return int32(C.nox_xxx_netSpriteUpdate_518AE0(asObjectC(recipient), C.int(player), (*C.uint32_t)(unsafe.Pointer(unit))))
}
