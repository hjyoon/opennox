package legacy

/*
#include "xfer_ammo_4f6b20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var ammoXferCall4F6B20 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerAmmoNative4F6B20(cf, object)
}

func ammoXferExportCall4F6B20(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerAmmo_4F6B20(asObjectC(object)))
}

//export nox_xxx_XFerAmmo_4F6B20
func nox_xxx_XFerAmmo_4F6B20(object *C.nox_object_t) C.int32_t {
	return C.int32_t(ammoXferCall4F6B20(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
