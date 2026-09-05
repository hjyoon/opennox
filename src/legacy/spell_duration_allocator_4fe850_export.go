package legacy

/*
#include "spell_duration_allocator_4fe850.h"
*/
import "C"

func spellDurationAllocatorLegacy4FE850() int32 {
	return GetServer().S().Spells.Dur.SpellCreateDurations4FE850()
}

func spellDurationAllocatorExportCall4FE850() int32 {
	return int32(C.nox_xxx_spellCreateDurations_4FE850())
}

//export nox_xxx_spellCreateDurations_4FE850
func nox_xxx_spellCreateDurations_4FE850() C.int32_t {
	return C.int32_t(spellDurationAllocatorLegacy4FE850())
}
