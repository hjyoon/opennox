package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_tradeAccept_50F5A0
func nox_xxx_tradeAccept_50F5A0(session unsafe.Pointer, actor *nox_object_t) {
	rt := game4TradeRuntime()
	rt.world.AcceptTradeNative50F5A0((*server.TradeSession)(session), asObjectS(actor), rt.p2p())
}

//export sub_50F790
func sub_50F790(session unsafe.Pointer, recipient *nox_object_t, head unsafe.Pointer) {
	rt := game4TradeRuntime()
	rt.world.TransferTradeOfferNative50F790((*server.TradeSession)(session), asObjectS(recipient), (*server.TradeItem)(head), rt.p2p())
}
