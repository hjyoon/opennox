package cfixture

/*
#include "GAME4_1.h"

typedef void (*shop_exit_t)(void*);
_Static_assert(_Generic(&nox_xxx_shopExit_50F4C0, shop_exit_t:1, default:0), "shop exit ABI");
static shop_exit_t volatile shop_exit_indirect = nox_xxx_shopExit_50F4C0;

static void shop_exit_fixture(int indirect, void* session) {
	if (indirect) {
		shop_exit_indirect(session);
	} else {
		nox_xxx_shopExit_50F4C0(session);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func ShopExitFixtureCall(route int, session *server.TradeSession) {
	C.shop_exit_fixture(C.int(route), unsafe.Pointer(session))
}
