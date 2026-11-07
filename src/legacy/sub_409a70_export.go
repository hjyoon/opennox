package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_409A70
func sub_409A70(value C.short) C.int {
	return C.int(sub409A70Native(int16(value)))
}
