package legacy

/*
#include "xfer_mover_4f5730.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var moverXferCall4F5730 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
	_ unsafe.Pointer,
) int32 {
	return Nox_xxx_XFerMoverNative4F5730(cf, object)
}

func moverXferExportCall4F5730(
	object *server.Object,
	context unsafe.Pointer,
) int32 {
	return int32(C.nox_xxx_XFerMover_4F5730(asObjectC(object), context))
}

//export nox_xxx_XFerMover_4F5730
func nox_xxx_XFerMover_4F5730(
	object *C.nox_object_t,
	context unsafe.Pointer,
) C.int32_t {
	return C.int32_t(moverXferCall4F5730(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
		context,
	))
}
