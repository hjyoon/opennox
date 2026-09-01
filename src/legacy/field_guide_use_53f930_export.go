package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
#include "server__magic__plyrgide.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func beastGuideAwardLegacy4FAE80(
	unit *server.Object,
	guide, notify int32,
) int32 {
	return GetServer().AwardBeastGuide4FAE80(unit, guide, notify)
}

func useFieldGuideLegacy53F930(owner, item *server.Object) int32 {
	return GetServer().UseFieldGuide53F930(owner, item)
}

func beastGuideAwardExportCall4FAE80(
	unit *server.Object,
	guide, notify int32,
) int32 {
	return int32(C.nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(
		asObjectC(unit),
		C.int32_t(guide),
		C.int32_t(notify),
	))
}

func fieldGuideUseExportCall53F930(owner, item *server.Object) int32 {
	return int32(C.sub_53F930(
		asObjectC(owner),
		asObjectC(item),
	))
}

// Nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide gives Go-owned callers the
// restored native-pointer service path.
func Nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(
	unit *server.Object,
	guide, notify int32,
) int32 {
	return beastGuideAwardLegacy4FAE80(unit, guide, notify)
}

//export nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide
func nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(
	unit *C.nox_object_t,
	guide, notify C.int32_t,
) C.int32_t {
	return C.int32_t(beastGuideAwardLegacy4FAE80(
		asObjectS((*nox_object_t)(unit)),
		int32(guide),
		int32(notify),
	))
}

//export sub_53F930
func sub_53F930(
	owner, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(useFieldGuideLegacy53F930(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
