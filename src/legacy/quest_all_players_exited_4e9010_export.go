package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export sub_4E9010
func sub_4E9010() C.int {
	return C.int(GetServer().S().QuestAllPlayersExited4E9010())
}
