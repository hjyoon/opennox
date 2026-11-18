package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_tradeP2PAddOfferMB_50FE20
func nox_xxx_tradeP2PAddOfferMB_50FE20(session unsafe.Pointer, code C.uint32_t) C.int {
	rt := game4TradeRuntime()
	ok := rt.world.RemoveTradeOfferNative50FE20((*server.TradeSession)(session), uint32(code), rt.p2p(), rt.reprice)
	return C.int(bool2int(ok))
}
