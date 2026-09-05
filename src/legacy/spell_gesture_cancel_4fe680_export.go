package legacy

/*
#include "spell_gesture_cancel_4fe680.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellGestureCancelLegacy4FE680(source *server.Object, radius float32) {
	GetServer().SpellGestureCancel4FE680(source, radius)
}

func spellGestureCancelExportCall4FE680(source *server.Object, radius float32) {
	C.nox_xxx_spell_4FE680(asObjectC(source), C.float(radius))
}

//export nox_xxx_spell_4FE680
func nox_xxx_spell_4FE680(source *C.nox_object_t, radius C.float) {
	spellGestureCancelLegacy4FE680(
		asObjectS((*nox_object_t)(source)),
		float32(radius),
	)
}
