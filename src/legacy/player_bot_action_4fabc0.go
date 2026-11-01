package legacy

/*
#include "defs.h"
*/
import "C"

//export nox_xxx_monsterActionToPlrState_4FABC0
func nox_xxx_monsterActionToPlrState_4FABC0(unit *nox_object_t) C.char {
	return C.char(asObjectS(unit).PlayerBotActionToState4FABC0())
}
