package legacy

/*
#include "god_mode_controller_4ef500.h"
*/
import "C"

var GodModeController4EF500 func(value uint32)

//export nox_xxx_set_god_4EF500
func nox_xxx_set_god_4EF500(value C.int32_t) {
	GodModeController4EF500(uint32(value))
}
