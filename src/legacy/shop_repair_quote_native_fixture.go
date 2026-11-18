package legacy

/*
#include "GAME4_1.h"

typedef uint32_t (*repair_quote_t)(nox_object_t*, void*, uint32_t);
_Static_assert(_Generic(&sub_5108D0, repair_quote_t:1, default:0), "repair quote ABI");

static repair_quote_t volatile repair_quote_indirect = sub_5108D0;

static uint32_t repair_quote_fixture(int indirect, nox_object_t* player, void* session, uint32_t code) {
	return indirect ? repair_quote_indirect(player, session, code) : sub_5108D0(player, session, code);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callShopRepairQuote5108D0(route int, player *server.Object, session *server.TradeSession, code uint32) uint32 {
	return uint32(C.repair_quote_fixture(C.int(route), asObjectC(player), unsafe.Pointer(session), C.uint32_t(code)))
}
