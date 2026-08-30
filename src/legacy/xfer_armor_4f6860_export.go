package legacy

/*
#include "xfer_armor_4f6860.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var armorXferCall4F6860 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerArmorNative4F6860(cf, object)
}

func armorXferExportCall4F6860(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerArmor_4F6860(asObjectC(object)))
}

//export nox_xxx_XFerArmor_4F6860
func nox_xxx_XFerArmor_4F6860(object *C.nox_object_t) C.int32_t {
	return C.int32_t(armorXferCall4F6860(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
