package legacy

/*
#include "defs.h"
#include "GAME2_3.h"
static void* ally_health_table_native(void) { return getMemAt(0x5D4594, 1200916); }
*/
import "C"

import "unsafe"

func clientAllyHealthTable495020() []byte {
	return unsafe.Slice((*byte)(C.ally_health_table_native()), 512)
}

func clientAllyHealthReset4951C0() unsafe.Pointer {
	return unsafe.Pointer(C.sub_4951C0())
}

func clientAllyHealthFree494FF0() unsafe.Pointer {
	return unsafe.Pointer(C.sub_494FF0())
}

func clientAllyHealthColor4950F0(code uint32, value byte) int {
	return int(C.sub_4950F0(C.int(int32(code)), C.char(value)))
}
