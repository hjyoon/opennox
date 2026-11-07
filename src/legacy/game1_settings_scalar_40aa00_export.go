package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_40AA00
func sub_40AA00() C.int { return C.int(sub40AA00Native()) }

//export sub_40AA20
func sub_40AA20() C.int { return C.int(sub40AA20Native()) }

//export sub_40AA30
func sub_40AA30(value C.int) C.int { return C.int(sub40AA30Native(int32(value))) }

//export sub_40AA40
func sub_40AA40() C.int { return C.int(sub40AA40Native()) }

//export sub_40AA60
func sub_40AA60(value C.int) C.int { return C.int(sub40AA60Native(int32(value))) }

//export sub_40E090
func sub_40E090() { sub40E090Native() }
