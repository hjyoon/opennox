package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_shopGetItemCost_50E3D0
func nox_xxx_shopGetItemCost_50E3D0(mode C.int, session unsafe.Pointer, item *nox_object_t) C.int {
	return C.int(game4TradeRuntime().world.ShopItemCostByModeNative50E3D0(int32(mode), (*server.TradeSession)(session), asObjectS(item)))
}
