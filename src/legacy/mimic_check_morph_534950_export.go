package legacy

/*
#include "defs.h"
void nox_xxx_monsterMimicCheckMorph_534950(nox_object_t* obj);
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var mimicCheckMorphCall534950 = func(obj *server.Object) {
	GetServer().S().MimicCheckMorph534950(obj)
}

//export nox_xxx_monsterMimicCheckMorph_native_534950
func nox_xxx_monsterMimicCheckMorph_native_534950(obj *C.nox_object_t) {
	mimicCheckMorphCall534950(asObjectS((*nox_object_t)(obj)))
}

func mimicCheckMorphCEntry534950(obj *server.Object) {
	C.nox_xxx_monsterMimicCheckMorph_534950(asObjectC(obj))
}
