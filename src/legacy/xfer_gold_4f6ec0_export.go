package legacy

/*
#include "xfer_gold_4f6ec0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var goldXferCall4F6EC0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerGoldNative4F6EC0(cf, object)
}

func goldXferExportCall4F6EC0(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerGold_4F6EC0(asObjectC(object)))
}

//export nox_xxx_XFerGold_4F6EC0
func nox_xxx_XFerGold_4F6EC0(object *C.nox_object_t) C.int32_t {
	return C.int32_t(goldXferCall4F6EC0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
