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

//export sub_50EEC0
func sub_50EEC0(node unsafe.Pointer) C.int {
	return C.int((*server.TradeItem)(node).ShopSortKey50EEC0())
}

//export sub_50F2B0
func sub_50F2B0(recipient *nox_object_t, node unsafe.Pointer) C.int {
	return C.int(game4TradeRuntime().world.SendShopItemNative50F2B0(asObjectS(recipient), (*server.TradeItem)(node)))
}

//export nox_xxx_servSendShopItems_50F280
func nox_xxx_servSendShopItems_50F280(recipient *nox_object_t, session unsafe.Pointer) C.uintptr_t {
	return C.uintptr_t(game4TradeRuntime().world.SendShopItemsNative50F280(asObjectS(recipient), (*server.TradeSession)(session)))
}
