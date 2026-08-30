package legacy

/*
#include "xfer_obelisk_4f6f60.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var obeliskXferCall4F6F60 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerObeliskNative4F6F60(cf, object)
}

func obeliskXferExportCall4F6F60(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerObelisk_4F6F60(asObjectC(object)))
}

//export nox_xxx_XFerObelisk_4F6F60
func nox_xxx_XFerObelisk_4F6F60(object *C.nox_object_t) C.int32_t {
	return C.int32_t(obeliskXferCall4F6F60(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
