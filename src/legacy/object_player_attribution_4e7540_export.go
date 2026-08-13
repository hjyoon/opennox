package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export sub_4E7540
func sub_4E7540(source, target *nox_object_t) {
	recordPlayerAttributionRuntime4E7540(asObjectS(source), asObjectS(target))
}
