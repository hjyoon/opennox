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
func sub_5109C0_trade(player *nox_object_t, session unsafe.Pointer, code C.uint32_t) C.uint32_t {
	rt := game4TradeRuntime()
	sell := rt.sell()
	var result uint32
	sell.SendQuote = func(player *server.Player, item *server.Object, cost uint32) {
		result = uint32(rt.world.SendShopSellQuoteNative5109C0(player, item, cost))
	}
	if rt.world.QuoteShopSellByCodeNative5109C0(asObjectS(player), (*server.TradeSession)(session), uint32(code), sell) == server.ShopSellNoItem5109C0 {
		return code
	}
	// The rejection path ends in a void audio call.
	return C.uint32_t(result)
}
