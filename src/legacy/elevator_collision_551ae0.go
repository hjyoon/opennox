package legacy

/*
#include "GAME5.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export sub_551AE0_go
func sub_551AE0_go(elevator, candidate *nox_object_t, candidateMoves C.int) {
	GetServer().S().ElevatorCollision551AE0(asObjectS(elevator), asObjectS(candidate), candidateMoves != 0,
		func(circle, box *server.Object, moveBox bool) {
			C.sub_54AD50(asObjectC(circle), asObjectC(box), C.int(bool2int(moveBox)))
		},
		func(first, second *server.Object) {
			C.sub_550F80(asObjectC(first), asObjectC(second))
		},
	)
}
