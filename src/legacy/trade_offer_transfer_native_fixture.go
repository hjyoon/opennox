package legacy

/*
#include "GAME4_1.h"

typedef void (*trade_offer_accept_t)(void*, nox_object_t*);
typedef void (*trade_offer_transfer_t)(void*, nox_object_t*, void*);
_Static_assert(_Generic(&nox_xxx_tradeAccept_50F5A0, trade_offer_accept_t:1, default:0), "trade accept ABI");
_Static_assert(_Generic(&sub_50F790, trade_offer_transfer_t:1, default:0), "trade transfer ABI");

static trade_offer_accept_t volatile trade_offer_accept_indirect = nox_xxx_tradeAccept_50F5A0;
static trade_offer_transfer_t volatile trade_offer_transfer_indirect = sub_50F790;

static void trade_offer_accept_fixture(int indirect, void* session, nox_object_t* actor) {
	if (indirect) trade_offer_accept_indirect(session, actor);
	else nox_xxx_tradeAccept_50F5A0(session, actor);
}

static void trade_offer_transfer_fixture(int indirect, void* session, nox_object_t* recipient, void* head) {
	if (indirect) trade_offer_transfer_indirect(session, recipient, head);
	else sub_50F790(session, recipient, head);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callTradeOfferAccept50F5A0(route int, session *server.TradeSession, actor *server.Object) {
	C.trade_offer_accept_fixture(C.int(route), unsafe.Pointer(session), asObjectC(actor))
}

func callTradeOfferTransfer50F790(route int, session *server.TradeSession, recipient *server.Object, head *server.TradeItem) {
	C.trade_offer_transfer_fixture(C.int(route), unsafe.Pointer(session), asObjectC(recipient), unsafe.Pointer(head))
}
