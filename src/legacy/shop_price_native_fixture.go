package legacy

/*
#include "GAME4_1.h"

typedef int (*shop_price_t)(int, void*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_shopGetItemCost_50E3D0, shop_price_t:1, default:0), "shop price ABI");

static shop_price_t volatile shop_price_indirect = nox_xxx_shopGetItemCost_50E3D0;

static int shop_price_fixture(int indirect, int mode, void* session, nox_object_t* item) {
	return indirect ? shop_price_indirect(mode, session, item) : nox_xxx_shopGetItemCost_50E3D0(mode, session, item);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callShopPriceNative50E3D0(route int, mode int32, session *server.TradeSession, item *server.Object) int32 {
	return int32(C.shop_price_fixture(C.int(route), C.int(mode), unsafe.Pointer(session), asObjectC(item)))
}
