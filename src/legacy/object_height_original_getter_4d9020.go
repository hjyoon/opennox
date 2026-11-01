package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
_Static_assert(_Generic(&nox_xxx_netReportUnitHeight_4D9020, int (*)(int,nox_object_t*):1,default:0), "public native two-word PE32 ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func heightOriginalCEntry4D9020(recipient int32, unit *server.Object) int32 {
	return int32(C.nox_xxx_netReportUnitHeight_4D9020(C.int(recipient), asObjectC(unit)))
}
