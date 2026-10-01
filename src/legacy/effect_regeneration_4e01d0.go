package legacy

/*
#include "defs.h"
void nox_xxx_effectRegenerationNative_4E01D0(void* effect, nox_object_t* item, void* context);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var regenerationEffectCall4E01D0 = func(effect *server.ModifierEff, item *server.Object) {
	GetServer().S().EffectRegeneration4E01D0(effect, item, unitAdjustHPCall4EE460)
}

func regenerationEffectPointerNative4E01D0() unsafe.Pointer {
	return C.nox_xxx_effectRegenerationNative_4E01D0
}

//export nox_xxx_effectRegenerationNative_4E01D0
func nox_xxx_effectRegenerationNative_4E01D0(effect unsafe.Pointer, item *C.nox_object_t, _ unsafe.Pointer) {
	regenerationEffectCall4E01D0((*server.ModifierEff)(effect), asObjectS((*nox_object_t)(item)))
}
