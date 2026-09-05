package legacy

/*
#include "spell_duration_free_recursive_4fe980.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellDurationFreeRecursiveLegacy4FE980(record unsafe.Pointer) {
	GetServer().S().Spells.Dur.SpellDurationFreeRecursive4FE980((*server.DurSpell)(record))
}

func spellDurationFreeRecursiveExportCall4FE980(record unsafe.Pointer) {
	C.sub_4FE980(record)
}

//export sub_4FE980
func sub_4FE980(record unsafe.Pointer) {
	spellDurationFreeRecursiveLegacy4FE980(record)
}
