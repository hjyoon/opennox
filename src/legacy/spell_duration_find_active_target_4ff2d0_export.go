package legacy

/*
#include "spell_duration_find_active_target_4ff2d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellDurationFindActiveTargetLegacy4FF2D0(
	spellID int32,
	target *server.Object,
) unsafe.Pointer {
	return unsafe.Pointer(
		GetServer().S().Spells.Dur.SpellDurationFindActiveTarget4FF2D0(spellID, target),
	)
}

func spellDurationFindActiveTargetExportCall4FF2D0(
	spellID int32,
	target *server.Object,
) unsafe.Pointer {
	return C.sub_4FF2D0(C.int32_t(spellID), asObjectC(target))
}

//export sub_4FF2D0
func sub_4FF2D0(spellID C.int32_t, target *C.nox_object_t) unsafe.Pointer {
	return spellDurationFindActiveTargetLegacy4FF2D0(
		int32(spellID),
		asObjectS((*nox_object_t)(target)),
	)
}
