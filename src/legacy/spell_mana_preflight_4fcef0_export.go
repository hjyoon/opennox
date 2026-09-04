package legacy

/*
#include "spell_mana_preflight_4fcef0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func spellManaPreflightExportCall4FCEF0(unit *server.Object, sequence *int32, count int32) int32 {
	return int32(C.nox_xxx_spellCheckSmth_4FCEF0(
		asObjectC(unit),
		(*C.int32_t)(unsafe.Pointer(sequence)),
		C.int32_t(count),
	))
}

//export nox_xxx_spellCheckSmth_4FCEF0
func nox_xxx_spellCheckSmth_4FCEF0(
	unit *C.nox_object_t,
	sequence *C.int32_t,
	count C.int32_t,
) C.int32_t {
	return C.int32_t(GetServer().S().SpellManaPreflight4FCEF0(
		asObjectS((*nox_object_t)(unit)),
		(*int32)(unsafe.Pointer(sequence)),
		int32(count),
	))
}
