package legacy

/*
#include "defs.h"
void nox_xxx_attribContinualReplen_4E02C0(void* effect, nox_object_t* item, void* context);
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func continualReplenishmentEffectPointer4E02C0() unsafe.Pointer {
	return C.nox_xxx_attribContinualReplen_4E02C0
}

//export nox_xxx_attribContinualReplen_4E02C0
func nox_xxx_attribContinualReplen_4E02C0(effect unsafe.Pointer, item *C.nox_object_t, _ unsafe.Pointer) {
	server.EffectContinualReplenishment4E02C0(
		(*server.ModifierEff)(effect), asObjectS((*nox_object_t)(item)),
		func() uint32 { return GetServer().S().Frame() },
		Nox_xxx_netReportCharges_4D82B0,
	)
}
