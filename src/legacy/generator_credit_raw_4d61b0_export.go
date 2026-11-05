package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
*/
import "C"

//export sub_4D61B0
func sub_4D61B0(unit *nox_object_t) {
	asObjectS(unit).RecordMonsterGeneratorCreditRaw4D61B0()
}
