package legacy

/*
#include "xfer_sentry_4f5e50.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var sentryXferCall4F5E50 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerSentryNative4F5E50(cf, object)
}

func sentryXferExportCall4F5E50(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerSentry_4F5E50(asObjectC(object)))
}

//export nox_xxx_XFerSentry_4F5E50
func nox_xxx_XFerSentry_4F5E50(object *C.nox_object_t) C.int32_t {
	return C.int32_t(sentryXferCall4F5E50(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
