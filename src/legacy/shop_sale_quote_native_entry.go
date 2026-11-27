package legacy

/*
#include "server__system__trade.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_5109C0_trade
func sub_5109C0_trade(player *nox_object_t, session unsafe.Pointer, code C.uint32_t) C.uintptr_t {
	rt := game4TradeRuntime()
	result := rt.world.QuoteShopSellByCodeNative5109C0(asObjectS(player), (*server.TradeSession)(session), uint32(code), rt.sell())
	return C.uintptr_t(result)
}
