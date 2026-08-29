package legacy

/*
#include "mover_update_54f740.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var moverUpdateCall54F740 = func(source *server.Object) {
	GetServer().S().MoverUpdate54F740(source, server.MoverUpdateRuntime54F740{
		Move: Nox_xxx_unitMove_4E7010,
	})
}

func moverUpdateExportCall54F740(source *server.Object) {
	C.nox_xxx_unitUpdateMover_54F740(asObjectC(source))
}

//export nox_xxx_unitUpdateMover_54F740
func nox_xxx_unitUpdateMover_54F740(source *C.nox_object_t) {
	moverUpdateCall54F740(asObjectS((*nox_object_t)(source)))
}
