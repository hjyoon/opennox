package legacy

/*
#include "GAME4_1.h"
#include "common/alloc/classes/alloc_class.h"

extern void* nox_alloc_tradeItems_2386496;

typedef void* (*shop_insert_native_t)(void*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_addItemToShopSession_50EE00, shop_insert_native_t:1, default:0), "shop insertion ABI");

static shop_insert_native_t volatile shop_insert_indirect = nox_xxx_addItemToShopSession_50EE00;

static void* shop_insert_test_call(int indirect, void* session, nox_object_t* object) {
	return indirect ? shop_insert_indirect(session, object) : nox_xxx_addItemToShopSession_50EE00(session, object);
}

static nox_alloc_class* shop_insert_test_pool(int count) {
	static char name[] = "ShopInsertionTestItems";
	return nox_new_alloc_class(name, 16, count);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func shopInsertionFixtureCall(route int, session *server.TradeSession, item *server.Object) *server.TradeItem {
	return (*server.TradeItem)(C.shop_insert_test_call(C.int(route), unsafe.Pointer(session), asObjectC(item)))
}

func shopInsertionFixtureUsePE32Pool(capacity int) (restore func()) {
	pool := C.shop_insert_test_pool(C.int(capacity))
	if pool == nil {
		return nil
	}
	previous := C.nox_alloc_tradeItems_2386496
	C.nox_alloc_tradeItems_2386496 = unsafe.Pointer(pool)
	return func() {
		C.nox_alloc_tradeItems_2386496 = previous
		C.nox_free_alloc_class(pool)
	}
}
