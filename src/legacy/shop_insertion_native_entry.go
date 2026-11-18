package legacy

/*
#include <string.h>
#include "GAME4_1.h"
#include "common/alloc/classes/alloc_class.h"

extern void* nox_alloc_tradeItems_2386496;

static void* nox_legacy_trade_alloc_item(void) {
	return nox_alloc_class_new_obj_zero((nox_alloc_class*)nox_alloc_tradeItems_2386496);
}

static uint32_t nox_legacy_shop_buy_cost(void* session, nox_object_t* item) {
	uint32_t address = (uint32_t)(uintptr_t)item;
	float argument;
	memcpy(&argument, &address, sizeof(address));
	return (uint32_t)nox_xxx_shopGetItemCost_50E3D0(1, (int)(uintptr_t)session, argument);
}
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
	node := (*server.TradeItem)(C.nox_legacy_trade_alloc_item())
	if node == nil {
		return nil
	}
	node.Item0 = item
	node.Cost4 = uint32(C.nox_legacy_shop_buy_cost(ptr, obj))
	session.InsertShopItem50EE00(node)
	return unsafe.Pointer(node)
}
