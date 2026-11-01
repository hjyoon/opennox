package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
_Static_assert(_Generic(&nox_xxx_netSendPhantomPlrMb_5187E0, int (*)(int,nox_object_t*):1,default:0), "native Phantom public ABI");
_Static_assert(_Generic(&nox_xxx_netSendSimpleObj_5188A0, int (*)(int,nox_object_t*):1,default:0), "native simple public ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func basicSenderOriginalCEntry5187E0(kind string, index int32, unit *server.Object) int32 {
	if kind == "phantom" {
		return int32(C.nox_xxx_netSendPhantomPlrMb_5187E0(C.int(index), asObjectC(unit)))
	}
	return int32(C.nox_xxx_netSendSimpleObj_5188A0(C.int(index), asObjectC(unit)))
}
