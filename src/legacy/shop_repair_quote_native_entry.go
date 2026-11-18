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
func sub_5108D0(player *nox_object_t, session unsafe.Pointer, code C.uint32_t) C.uint32_t {
	rt := game4TradeRuntime()
	repair := rt.repair()
	var result uint32
	repair.SendQuote = func(player *server.Player, item *server.Object, cost uint32) {
		result = uint32(rt.world.SendShopRepairQuoteNative5108D0(player, item, cost))
	}
	// The reject branch calls a void audio service. Only the quote branch
	// has a defined packet-queue return value.
	rt.world.QuoteShopRepairByCodeNative5108D0(asObjectS(player), (*server.TradeSession)(session), uint32(code), repair)
	return C.uint32_t(result)
}
