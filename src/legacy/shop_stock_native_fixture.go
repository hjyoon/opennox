package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"

typedef int (*shop_sort_key_t)(void*);
typedef int (*shop_stock_item_t)(nox_object_t*, void*);
typedef uintptr_t (*shop_stock_list_t)(nox_object_t*, void*);
_Static_assert(_Generic(&sub_50EEC0, shop_sort_key_t:1, default:0), "shop sort key ABI");
_Static_assert(_Generic(&sub_50F2B0, shop_stock_item_t:1, default:0), "shop item packet ABI");
_Static_assert(_Generic(&nox_xxx_servSendShopItems_50F280, shop_stock_list_t:1, default:0), "shop stock list ABI");
_Static_assert(sizeof(int)==4, "shop queue result width");

static shop_sort_key_t volatile shop_sort_key_indirect = sub_50EEC0;
static shop_stock_item_t volatile shop_stock_item_indirect = sub_50F2B0;
static shop_stock_list_t volatile shop_stock_list_indirect = nox_xxx_servSendShopItems_50F280;

static uintptr_t shop_stock_call(int entry, int indirect, nox_object_t* recipient, void* node, void* session) {
	switch (entry) {
	case 0:
		return (uint32_t)(indirect ? shop_sort_key_indirect(node) : sub_50EEC0(node));
	case 1:
		return (uint32_t)(indirect ? shop_stock_item_indirect(recipient, node) : sub_50F2B0(recipient, node));
	default:
		return indirect ? shop_stock_list_indirect(recipient, session) : nox_xxx_servSendShopItems_50F280(recipient, session);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func shopStockFixtureCall(entry, route int, recipient *server.Object, session *server.TradeSession) uintptr {
	return uintptr(C.shop_stock_call(C.int(entry), C.int(route), asObjectC(recipient),
		unsafe.Pointer(session.Field20), unsafe.Pointer(session)))
}
