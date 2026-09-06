package legacy

/*
#include "spell_duration_item_cancel_4feb60.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func itemCancelDurSpellsLegacy4FEB60(owner, item *server.Object) {
	GetServer().S().Spells.Dur.ItemCancelDurSpells4FEB60(owner, item)
}

func itemCancelDurSpellsExportCall4FEB60(owner, item *server.Object) {
	C.sub_4FEB60(asObjectC(owner), asObjectC(item))
}

//export sub_4FEB60
func sub_4FEB60(owner, item *C.nox_object_t) {
	itemCancelDurSpellsLegacy4FEB60(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	)
}
