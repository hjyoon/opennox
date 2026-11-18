package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_createShopStruct_50E870
func nox_xxx_createShopStruct_50E870() unsafe.Pointer {
	return unsafe.Pointer(game4TradeRuntime().world.NewTradeSessionNative50E870())
}

//export nox_xxx_createPlayerShopSession_50E8F0
func nox_xxx_createPlayerShopSession_50E8F0(player, merchant *nox_object_t) unsafe.Pointer {
	return unsafe.Pointer(game4TradeRuntime().open(asObjectS(player), asObjectS(merchant)))
}

//export nox_xxx_loadShopItems_50E970
func nox_xxx_loadShopItems_50E970(session unsafe.Pointer) {
	game4TradeRuntime().load((*server.TradeSession)(session))
}
