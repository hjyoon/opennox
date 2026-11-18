package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export nox_xxx_useByNetCode_53F8E0
func nox_xxx_useByNetCode_53F8E0(owner, item *nox_object_t) C.int {
	object := asObjectS(item)
	if object == nil || object.Use.Ptr == nil || sub_419E60(owner) == 1 {
		return 1
	}
	return C.int(ccall.CallIntPtr2(object.Use.Ptr, unsafe.Pointer(owner), unsafe.Pointer(item)))
}

//export nox_xxx_mobActionPickupObject_544B90
func nox_xxx_mobActionPickupObject_544B90(unit *nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionPickupObject544B90(asObjectS(unit), Nox_xxx_inventoryServPlace_4F36F0))
}
