package legacy

/*
#include "GAME4.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerScheduledSpellExportCall4FB0E0(unit, target *server.Object) int32 {
	return int32(C.nox_xxx_playerDoSchedSpell_4FB0E0(asObjectC(unit), asObjectC(target)))
}

func playerScheduledSpellQueueExportCall4FB1D0(unit, target *server.Object) int32 {
	return int32(C.nox_xxx_playerDoSchedSpellQueue_4FB1D0(asObjectC(unit), asObjectC(target)))
}

//export nox_xxx_playerDoSchedSpell_4FB0E0
func nox_xxx_playerDoSchedSpell_4FB0E0(unit, target *C.nox_object_t) C.int {
	return C.int(playerDoScheduledSpellLegacy4FB0E0(
		asObjectS((*nox_object_t)(unit)),
		asObjectS((*nox_object_t)(target)),
	))
}

//export nox_xxx_playerDoSchedSpellQueue_4FB1D0
func nox_xxx_playerDoSchedSpellQueue_4FB1D0(unit, target *C.nox_object_t) C.int {
	return C.int(playerDoScheduledSpellQueueLegacy4FB1D0(
		asObjectS((*nox_object_t)(unit)),
		asObjectS((*nox_object_t)(target)),
	))
}
