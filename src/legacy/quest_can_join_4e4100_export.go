package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export sub_4E4100
func sub_4E4100() C.int {
	return C.int(GetServer().S().QuestCanJoin4E4100())
}
