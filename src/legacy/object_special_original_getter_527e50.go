package legacy

/*
#include "defs.h"
#include "GAME4_2.h"
_Static_assert(_Generic(&nox_xxx_netUpdateObjectSpecial_527E50, int (*)(nox_object_t*,nox_object_t*):1,default:0), "native public two-pointer ABI");
*/
import "C"
import "github.com/opennox/opennox/v1/server"

func specialFlagsOriginalCEntry527E50(r, u *server.Object) int32 {
	return int32(C.nox_xxx_netUpdateObjectSpecial_527E50(asObjectC(r), asObjectC(u)))
}
