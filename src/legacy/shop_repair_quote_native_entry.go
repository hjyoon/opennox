package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_5108D0
func sub_5108D0(player *nox_object_t, session unsafe.Pointer, code C.uint32_t) C.uintptr_t {
	rt := game4TradeRuntime()
	result := rt.world.QuoteShopRepairByCodeNative5108D0(asObjectS(player), (*server.TradeSession)(session), uint32(code), rt.repair())
	return C.uintptr_t(result)
}
