package legacy

/*
#include "GAME1.h"
*/
import "C"

import "unsafe"

//export sub_40AA70
func sub_40AA70(player *C.nox_playerInfo) C.int {
	return C.int(sub40AA70Native(unsafe.Pointer(player)))
}
