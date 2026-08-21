package legacy

/*
#include "break_init_4f0570.h"
*/
import "C"

//export nox_xxx_breakInit_4F0570
func nox_xxx_breakInit_4F0570(unit *C.nox_object_t) {
	breakInitCall4F0570(asObjectS((*nox_object_t)(unit)))
}
