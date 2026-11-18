package legacy

/*
#include "GAME4_1.h"

typedef void (*cancel_session_t)(void*);
typedef void (*clear_shop_cache_t)(int);
_Static_assert(_Generic(&nox_xxx_shopCancelSession_510DC0, cancel_session_t:1, default:0), "trade cancel ABI");
_Static_assert(_Generic(&sub_510E20, clear_shop_cache_t:1, default:0), "quest shop cache ABI");

static cancel_session_t volatile cancel_session_indirect = nox_xxx_shopCancelSession_510DC0;
static clear_shop_cache_t volatile clear_shop_cache_indirect = sub_510E20;

static void cancel_session_probe(int indirect, void* session) {
	if (indirect) cancel_session_indirect(session);
	else nox_xxx_shopCancelSession_510DC0(session);
}

static void clear_shop_cache_probe(int indirect, int index) {
	if (indirect) clear_shop_cache_indirect(index);
	else sub_510E20(index);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callTradeCancel510DC0(route int, session *server.TradeSession) {
	C.cancel_session_probe(C.int(route), unsafe.Pointer(session))
}

func callQuestCacheClear510E20(route, index int) {
	C.clear_shop_cache_probe(C.int(route), C.int(index))
}
