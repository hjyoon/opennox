package legacy

/*
#include "GAME3_2.h"

typedef uint32_t* (*shadow_object_fn)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_unitNewAddShadow_4DA9A0, shadow_object_fn:1, default:0), "add shadow ABI");
_Static_assert(_Generic(&nox_xxx_action_4DA9F0, shadow_object_fn:1, default:0), "remove shadow ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func addObjectShadowC(obj *server.Object) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_unitNewAddShadow_4DA9A0(asObjectC(obj)))
}

func removeObjectShadowC(obj *server.Object) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_action_4DA9F0(asObjectC(obj)))
}
