package legacy

/*
#include "GAME4_1.h"

typedef int (*offer_remove_t)(void*, uint32_t);
_Static_assert(_Generic(&nox_xxx_tradeP2PAddOfferMB_50FE20, offer_remove_t:1, default:0), "offer removal ABI");

static offer_remove_t volatile offer_remove_indirect = nox_xxx_tradeP2PAddOfferMB_50FE20;

static int offer_remove_fixture(int indirect, void* session, uint32_t code) {
	return indirect ? offer_remove_indirect(session, code) : nox_xxx_tradeP2PAddOfferMB_50FE20(session, code);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callTradeOfferRemove50FE20(route int, session *server.TradeSession, code uint32) int {
	return int(C.offer_remove_fixture(C.int(route), unsafe.Pointer(session), C.uint32_t(code)))
}
