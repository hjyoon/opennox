package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_40A740
func sub_40A740() C.int { return C.int(sub40A740Native()) }

//export sub_40A770
func sub_40A770() C.int { return C.int(sub40A770Native()) }
