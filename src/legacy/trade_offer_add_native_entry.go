package legacy

/*
#include "server__system__trade.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_50FD60
func sub_50FD60(head unsafe.Pointer, item *nox_object_t) C.int {
	return C.int(bool2int(server.TradeOfferItemAllowed50FD60((*server.TradeItem)(head), asObjectS(item))))
}

//export nox_xxx_tradeP2PAddOffer2_50F820_trade
func nox_xxx_tradeP2PAddOffer2_50F820_trade(session unsafe.Pointer, actor, item *nox_object_t) C.int {
	rt := game4TradeRuntime()
	ok := rt.world.AddTradeOfferNative50F820((*server.TradeSession)(session), asObjectS(actor), asObjectS(item), rt.p2p(), rt.reprice)
	return C.int(bool2int(ok))
}
