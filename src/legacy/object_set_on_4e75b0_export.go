package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export nox_xxx_objectSetOn_4E75B0
func nox_xxx_objectSetOn_4E75B0(obj *nox_object_t) C.char {
	return C.char(objectSetOnRuntime4E75B0(asObjectS(obj)))
}
