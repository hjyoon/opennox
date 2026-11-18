package legacy

/*
#include "server__system__trade.h"

typedef int (*trade_start_packet_t)(nox_object_t*, void*);
typedef void* (*trade_start_t)(nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&sub_50F0F0, trade_start_packet_t:1, default:0), "shop start packet ABI");
_Static_assert(_Generic(&sub_50F1A0, trade_start_packet_t:1, default:0), "trade start packet ABI");
_Static_assert(_Generic(&nox_xxx_servShopStart_50EF10_trade, trade_start_t:1, default:0), "trade start ABI");

static trade_start_packet_t volatile shop_start_packet_indirect = sub_50F0F0;
static trade_start_packet_t volatile trade_start_packet_indirect = sub_50F1A0;
static trade_start_t volatile trade_start_indirect = nox_xxx_servShopStart_50EF10_trade;

static int trade_start_packet_fixture(int shop, int indirect, nox_object_t* recipient, void* session) {
	if (shop) {
		return indirect ? shop_start_packet_indirect(recipient, session) : sub_50F0F0(recipient, session);
	}
	return indirect ? trade_start_packet_indirect(recipient, session) : sub_50F1A0(recipient, session);
}

static void* trade_start_fixture(int indirect, nox_object_t* starter, nox_object_t* target) {
	return indirect ? trade_start_indirect(starter, target) : nox_xxx_servShopStart_50EF10_trade(starter, target);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func tradeStartPacketFixtureCall(shop, route int, recipient *server.Object, session *server.TradeSession) uint32 {
	return uint32(C.trade_start_packet_fixture(C.int(shop), C.int(route), asObjectC(recipient), unsafe.Pointer(session)))
}

func tradeStartFixtureCall(route int, starter, target *server.Object) *server.TradeSession {
	return (*server.TradeSession)(C.trade_start_fixture(C.int(route), asObjectC(starter), asObjectC(target)))
}
