package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
typedef char (*dispatch_519410_fn)(nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_netSendObjects2Plr_519410, dispatch_519410_fn: 1, default: 0), "519410 pointer and byte ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func objectDispatchPublicC519410(recipient, unit *server.Object) byte {
	return byte(C.nox_xxx_netSendObjects2Plr_519410(asObjectC(recipient), asObjectC(unit)))
}
