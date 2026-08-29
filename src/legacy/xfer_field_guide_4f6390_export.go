package legacy

/*
#include "xfer_field_guide_4f6390.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var fieldGuideXferCall4F6390 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerFieldGuideNative4F6390(cf, object)
}

func fieldGuideXferExportCall4F6390(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerFieldGuide_4F6390(asObjectC(object)))
}

//export nox_xxx_XFerFieldGuide_4F6390
func nox_xxx_XFerFieldGuide_4F6390(object *C.nox_object_t) C.int32_t {
	return C.int32_t(fieldGuideXferCall4F6390(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
