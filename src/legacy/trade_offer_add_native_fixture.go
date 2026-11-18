package legacy

/*
#include "server__system__trade.h"

typedef int (*offer_eligible_t)(void*, nox_object_t*);
typedef int (*offer_add_t)(void*, nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&sub_50FD60, offer_eligible_t:1, default:0), "offer eligibility ABI");
_Static_assert(_Generic(&nox_xxx_tradeP2PAddOffer2_50F820_trade, offer_add_t:1, default:0), "offer add ABI");

static offer_eligible_t volatile offer_eligible_indirect = sub_50FD60;
static offer_add_t volatile offer_add_indirect = nox_xxx_tradeP2PAddOffer2_50F820_trade;

static int offer_eligible_fixture(int indirect, void* head, nox_object_t* item) {
	return indirect ? offer_eligible_indirect(head, item) : sub_50FD60(head, item);
}

static int offer_add_fixture(int indirect, void* session, nox_object_t* actor, nox_object_t* item) {
	return indirect ? offer_add_indirect(session, actor, item) : nox_xxx_tradeP2PAddOffer2_50F820_trade(session, actor, item);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callTradeOfferEligible50FD60(route int, head *server.TradeItem, item *server.Object) int {
	return int(C.offer_eligible_fixture(C.int(route), unsafe.Pointer(head), asObjectC(item)))
}

func callTradeOfferAdd50F820(route int, session *server.TradeSession, actor, item *server.Object) int {
	return int(C.offer_add_fixture(C.int(route), unsafe.Pointer(session), asObjectC(actor), asObjectC(item)))
}
