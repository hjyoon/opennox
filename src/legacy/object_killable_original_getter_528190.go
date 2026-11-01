package legacy

/*
#include "defs.h"
#include "GAME4_2.h"
_Static_assert(_Generic(&nox_xxx_checkIsKillable_528190, int (*)(nox_object_t*):1,default:0), "public native pointer ABI");
*/
import "C"
import "github.com/opennox/opennox/v1/server"

func killableOriginalCEntry528190(u *server.Object) int32 {
	return int32(C.nox_xxx_checkIsKillable_528190(asObjectC(u)))
}
