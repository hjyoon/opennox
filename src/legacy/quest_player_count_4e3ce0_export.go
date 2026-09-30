package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export nox_xxx_questPlayerCount_native_4E3CE0
func nox_xxx_questPlayerCount_native_4E3CE0() C.int {
	return C.int(GetServer().S().QuestPlayerCount4E3CE0())
}
