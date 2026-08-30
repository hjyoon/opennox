package legacy

/*
#include "server__magic__plyrspel.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellGrantToPlayerLegacy4FB550(
	unit *server.Object,
	spellID, notify, shop, override int32,
) int32 {
	return GetServer().SpellGrantToPlayer4FB550(unit, spellID, notify, shop, override)
}

func spellGrantExportCall4FB550(
	unit *server.Object,
	spellID, notify, shop, override int32,
) int32 {
	return int32(C.nox_xxx_spellGrantToPlayer_4FB550(
		asObjectC(unit),
		C.int(spellID),
		C.int(notify),
		C.int(shop),
		C.int(override),
	))
}

//export nox_xxx_spellGrantToPlayer_4FB550
func nox_xxx_spellGrantToPlayer_4FB550(
	unit *C.nox_object_t,
	spellID, notify, shop, override C.int,
) C.int {
	return C.int(spellGrantToPlayerLegacy4FB550(
		asObjectS((*nox_object_t)(unit)),
		int32(spellID),
		int32(notify),
		int32(shop),
		int32(override),
	))
}
