package legacy

/*
#include "spell_duration_unlink_4fe900.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellDurationUnlinkLegacy4FE900(record unsafe.Pointer) {
	GetServer().S().Spells.Dur.SpellDurationUnlink4FE900((*server.DurSpell)(record))
}

func spellDurationUnlinkExportCall4FE900(record unsafe.Pointer) {
	C.sub_4FE900(record)
}

//export sub_4FE900
func sub_4FE900(record unsafe.Pointer) {
	spellDurationUnlinkLegacy4FE900(record)
}
