package legacy

/*
#include "xfer_weapon_4f64a0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var weaponXferCall4F64A0 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerWeaponNative4F64A0(cf, object)
}

func weaponXferExportCall4F64A0(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerWeapon_4F64A0(asObjectC(object)))
}

//export nox_xxx_XFerWeapon_4F64A0
func nox_xxx_XFerWeapon_4F64A0(object *C.nox_object_t) C.int32_t {
	return C.int32_t(weaponXferCall4F64A0(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
