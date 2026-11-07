package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_40A1F0
func sub_40A1F0(argument C.int) C.int { return C.int(sub40A1F0Native(int32(argument))) }

//export sub_40A230
func sub_40A230() C.int { return C.int(sub40A230Native()) }

//export sub_40A250
func sub_40A250() C.longlong { return C.longlong(sub40A250Native()) }

//export sub_40A310
func sub_40A310(argument C.int) C.longlong { return C.longlong(sub40A310Native(int32(argument))) }
