package cfixture

/*
#include "GAME4_1.h"
#include "server__system__trade.h"


typedef void (*trade_release_t)(void*);
typedef int (*offer_release_t)(void*);
_Static_assert(_Generic(&sub_510000, trade_release_t:1, default:0), "trade release ABI");
_Static_assert(_Generic(&sub_50F6B0, offer_release_t:1, default:0), "offer release ABI");

static trade_release_t volatile fixture_trade_release = sub_510000;
static offer_release_t volatile fixture_offer_release = sub_50F6B0;

static void trade_release_fixture_call(int indirect, void* session) {
	if (indirect) {
		fixture_trade_release(session);
	} else {
		sub_510000(session);
	}
}

static int offer_release_fixture_call(int indirect, void* head) {
	return indirect ? fixture_offer_release(head) : sub_50F6B0(head);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TradeReleaseFixtureCall(route int, session *server.TradeSession) {
	C.trade_release_fixture_call(C.int(route), unsafe.Pointer(session))
}

func TradeOfferReleaseFixtureCall(route int, head *server.TradeItem) int {
	return int(C.offer_release_fixture_call(C.int(route), unsafe.Pointer(head)))
}

func TradeReleaseFixtureUsePE32Pools(sessions, items *alloc.Class, head *server.TradeSession) func() {
	state := C.nox_legacy_trade_pool_storage()
	oldSessions, oldItems, oldHead := state.sessions, state.items, state.head_word
	state.sessions = sessions.UPtr()
	state.items = items.UPtr()
	state.head_word = C.uint32_t(uintptr(unsafe.Pointer(head)))
	return func() {
		state.sessions, state.items, state.head_word = oldSessions, oldItems, oldHead
	}
}

func TradeReleaseFixturePE32Head() uintptr {
	return uintptr(C.nox_legacy_trade_pool_storage().head_word)
}
