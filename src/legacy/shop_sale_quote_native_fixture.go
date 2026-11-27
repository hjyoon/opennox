package legacy

/*
#include "server__system__trade.h"

typedef uintptr_t (*sale_quote_t)(nox_object_t*, void*, uint32_t);
_Static_assert(_Generic(&sub_5109C0_trade, sale_quote_t:1, default:0), "sale quote ABI");

static sale_quote_t volatile sale_quote_indirect = sub_5109C0_trade;

static uintptr_t sale_quote_fixture(int indirect, nox_object_t* player, void* session, uint32_t code) {
	return indirect ? sale_quote_indirect(player, session, code) : sub_5109C0_trade(player, session, code);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callShopSaleQuote5109C0(route int, player *server.Object, session *server.TradeSession, code uint32) uintptr {
	return uintptr(C.sale_quote_fixture(C.int(route), asObjectC(player), unsafe.Pointer(session), C.uint32_t(code)))
}
