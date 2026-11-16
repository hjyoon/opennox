package legacy

/*
#include "GAME4_2.h"
*/
import "C"

//export sub_52BA70
func sub_52BA70(obj *C.nox_object_t) C.int {
	npcNormalizeEquipped52BA70(asObjectS((*nox_object_t)(obj)))
	// Every returning original path ends with a null inventory cursor.
	return 0
}
