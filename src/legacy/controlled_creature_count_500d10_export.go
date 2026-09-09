package legacy

/*
#include "controlled_creature_count_500d10.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var controlledCreatureCountCall500D10 = func(owner *server.Object) int32 {
	return int32(owner.Nox_xxx_countControlledCreatures_500D10())
}

func controlledCreatureCountExportCall500D10(owner *server.Object) int32 {
	return int32(C.nox_xxx_countControlledCreatures_500D10(asObjectC(owner)))
}

//export nox_xxx_countControlledCreatures_500D10
func nox_xxx_countControlledCreatures_500D10(owner *C.nox_object_t) C.int32_t {
	return C.int32_t(controlledCreatureCountCall500D10(
		asObjectS((*nox_object_t)(owner)),
	))
}
