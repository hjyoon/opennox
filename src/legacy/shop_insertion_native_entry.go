package legacy

/*
#include "GAME4_1.h"
#include "common/alloc/classes/alloc_class.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_addItemToShopSession_50EE00
func nox_xxx_addItemToShopSession_50EE00(ptr unsafe.Pointer, obj *nox_object_t) unsafe.Pointer {
	session, item := (*server.TradeSession)(ptr), asObjectS(obj)
	if session == nil || item == nil {
		return nil
	}
	world := game4TradeRuntime().world
	if world.IsTradeSessionNative(session) {
		return unsafe.Pointer(world.AddShopItemNative50EE00(session, item))
	}
	// Sessions from the remaining PE32 allocator use the same record layout
	// as TradeSession on 32-bit hosts. Their nodes remain owned by that pool.
	if unsafe.Sizeof(uintptr(0)) != 4 {
		return nil
	}
	node := (*server.TradeItem)(nox_alloc_class_new_obj_zero((*C.nox_alloc_class)(legacyTradePools.items)))
	if node == nil {
		return nil
	}
	node.Item0 = item
	node.Cost4 = uint32(world.ShopItemCostByModeNative50E3D0(1, session, item))
	session.InsertShopItem50EE00(node)
	return unsafe.Pointer(node)
}
