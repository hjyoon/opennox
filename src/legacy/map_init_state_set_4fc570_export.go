package legacy

/*
#include <stdint.h>

#include "map_init_state_set_4fc570.h"
*/
import "C"

func mapInitStateSetExportCall4FC570(value int32) int32 {
	return int32(C.nox_xxx_resetMapInit_4FC570(C.int32_t(value)))
}

//export nox_xxx_resetMapInit_4FC570
func nox_xxx_resetMapInit_4FC570(value C.int32_t) C.int32_t {
	return C.int32_t(GetServer().S().SetMapInitState4FC570(int32(value)))
}
