package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
void nox_xxx_getMissilesInCircle_518170(float2*,float,void*,nox_object_t*);
static inline void nox_anti_spell_probe_53BB00(int kind,nox_object_t* missile,nox_object_t* candidate) {
	if (kind) sub_53BD10(candidate,missile);
	else nox_xxx_updateAntiSpellProj_53BB00(missile);
}
static inline void nox_anti_spell_circle_probe_518170(float2* origin,nox_object_t* missile) {
 nox_xxx_getMissilesInCircle_518170(origin,600.0,(void*)sub_53BD10,missile);
}
static inline void nox_anti_spell_nil_circle_probe_518170(void) {
 nox_xxx_getMissilesInCircle_518170(0,600.0,0,0);
}
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func antiSpellCProbe53BB00(kind int, missile, candidate *server.Object) {
	C.nox_anti_spell_probe_53BB00(C.int(kind), asObjectC(missile), asObjectC(candidate))
}

func antiSpellCircleCProbe518170(origin *types.Pointf, missile *server.Object) {
	C.nox_anti_spell_circle_probe_518170((*C.float2)(unsafe.Pointer(origin)), asObjectC(missile))
}

func antiSpellNilCircleCProbe518170() { C.nox_anti_spell_nil_circle_probe_518170() }
