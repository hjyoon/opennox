package legacy

/*
#include "pickup_spellbook_4f3c60.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupSpellBookCall4F3C60(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupSpellbook_4F3C60(owner, item, arg3, arg4)
}

func pickupSpellBookExportCall4F3C60(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupSpellbook_4F3C60(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

//export nox_xxx_pickupSpellbook_4F3C60
func nox_xxx_pickupSpellbook_4F3C60(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupSpellBookCall4F3C60(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
