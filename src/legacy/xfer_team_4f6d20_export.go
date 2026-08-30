package legacy

/*
#include "xfer_team_4f6d20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var teamXferCall4F6D20 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerTeamNative4F6D20(cf, object)
}

func teamXferExportCall4F6D20(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerTeam_4F6D20(asObjectC(object)))
}

//export nox_xxx_XFerTeam_4F6D20
func nox_xxx_XFerTeam_4F6D20(object *C.nox_object_t) C.int32_t {
	return C.int32_t(teamXferCall4F6D20(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
