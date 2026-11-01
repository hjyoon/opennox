package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&nox_xxx_netPlayerObjSend_518C30, int (*)(nox_object_t*,nox_object_t*,int,int):1, default:0), "player sender original ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerSendOriginalPublicC518C30(recipient, unit *server.Object, stream, health int32) int32 {
	return int32(C.nox_xxx_netPlayerObjSend_518C30(asObjectC(recipient), asObjectC(unit), C.int(stream), C.int(health)))
}
