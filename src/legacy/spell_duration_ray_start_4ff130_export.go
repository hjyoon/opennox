package legacy

/*
#include "spell_duration_ray_start_4ff130.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func durationRayStartLegacy4FF130(record unsafe.Pointer) {
	GetServer().S().DurationRayStart4FF130((*server.DurSpell)(record))
}

func durationRayStartExportCall4FF130(record unsafe.Pointer) {
	C.nox_xxx_netStartDurationRaySpell_4FF130(record)
}

//export nox_xxx_netStartDurationRaySpell_4FF130
func nox_xxx_netStartDurationRaySpell_4FF130(record unsafe.Pointer) {
	durationRayStartLegacy4FF130(record)
}
