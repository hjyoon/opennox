package legacy

/*
#include "GAME3_3.h"
*/
import "C"

//export nox_xxx_questCanJoin_native_4E4100
func nox_xxx_questCanJoin_native_4E4100() C.uint {
	return C.uint(GetServer().S().QuestCanJoin4E4100())
}
