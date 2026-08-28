package legacy

/*
#include "xfer_spell_page_pedestal_4f4a20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

//export nox_xxx_XFerSpellPagePedistal_4F4A20
func nox_xxx_XFerSpellPagePedistal_4F4A20(
	object *C.nox_object_t,
	_ unsafe.Pointer,
) C.int32_t {
	return C.int32_t(Nox_xxx_XFerSpellPagePedestalNative4F4A20(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
