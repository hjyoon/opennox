package legacy

/*
#include "spell_duration_cleanup_4fe880.h"
*/
import "C"

func spellDurationCleanupLegacy4FE880() {
	GetServer().S().Spells.Dur.SpellFreeDurations4FE880()
}

func spellDurationCleanupExportCall4FE880() {
	C.sub_4FE880()
}

//export sub_4FE880
func sub_4FE880() {
	spellDurationCleanupLegacy4FE880()
}
