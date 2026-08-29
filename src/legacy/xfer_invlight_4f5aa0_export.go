package legacy

/*
#include "xfer_invlight_4f5aa0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var invLightXferCall4F5AA0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerInvLightNative4F5AA0(cf, object)
}

func invLightXferExportCall4F5AA0(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerInvLight_4F5AA0(asObjectC(object)))
}

//export nox_xxx_XFerInvLight_4F5AA0
func nox_xxx_XFerInvLight_4F5AA0(object *C.nox_object_t) C.int32_t {
	return C.int32_t(invLightXferCall4F5AA0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
