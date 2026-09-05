package legacy

/*
#include "spell_duration_selective_cleanup_4fe8a0.h"
*/
import "C"

func spellDurationSelectiveCleanupLegacy4FE8A0(mode int32) {
	GetServer().S().Spells.Dur.SpellResetDurations4FE8A0(mode)
}

func spellDurationSelectiveCleanupExportCall4FE8A0(mode int32) {
	C.sub_4FE8A0(C.int32_t(mode))
}

//export sub_4FE8A0
func sub_4FE8A0(mode C.int32_t) {
	spellDurationSelectiveCleanupLegacy4FE8A0(int32(mode))
}
