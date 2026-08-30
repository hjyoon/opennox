package legacy

/*
#include "xfer_toxic_cloud_4f70a0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var toxicCloudXferCall4F70A0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerToxicCloudNative4F70A0(cf, object)
}

func toxicCloudXferExportCall4F70A0(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerToxicCloud_4F70A0(asObjectC(object)))
}

//export nox_xxx_XFerToxicCloud_4F70A0
func nox_xxx_XFerToxicCloud_4F70A0(object *C.nox_object_t) C.int32_t {
	return C.int32_t(toxicCloudXferCall4F70A0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
