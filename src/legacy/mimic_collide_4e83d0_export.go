package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "unsafe"

//export nox_xxx_collideMimic_4E83D0
func nox_xxx_collideMimic_4E83D0(mimic, other *C.nox_object_t, collision *C.float) unsafe.Pointer {
	srv := GetServer()
	return srv.S().MimicCollide4E83D0(
		asObjectS((*nox_object_t)(mimic)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
		srv.NoxScriptC().ScriptCallback,
	)
}
