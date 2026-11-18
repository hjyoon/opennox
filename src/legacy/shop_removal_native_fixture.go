package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"

typedef int (*shop_remove_callback)(nox_object_t*, nox_object_t*);
typedef int (*trade_close_callback)(nox_object_t*);
typedef int (*shop_gold_callback)(nox_object_t*, short);
typedef int (*shop_unlink_callback)(void*, nox_object_t*);
_Static_assert(_Generic(&sub_50E820, shop_remove_callback:1, default:0), "shop removal packet ABI");
_Static_assert(_Generic(&sub_50F450, trade_close_callback:1, default:0), "trade close ABI");
_Static_assert(_Generic(&sub_5104F0, shop_gold_callback:1, default:0), "missing gold ABI");
_Static_assert(_Generic(&sub_50E7A0, shop_unlink_callback:1, default:0), "shop unlink ABI");
_Static_assert(sizeof(short)==2 && sizeof(int)==4, "packet scalar widths");

static shop_remove_callback volatile shop_remove_indirect = sub_50E820;
static trade_close_callback volatile trade_close_indirect = sub_50F450;
static shop_gold_callback volatile shop_gold_indirect = sub_5104F0;
static shop_unlink_callback volatile shop_unlink_indirect = sub_50E7A0;

static int shop_removal_call(int entry, int indirect, nox_object_t* recipient, nox_object_t* item, void* session, short amount) {
	switch (entry) {
	case 0:
		return indirect ? shop_remove_indirect(recipient, item) : sub_50E820(recipient, item);
	case 1:
		return indirect ? trade_close_indirect(recipient) : sub_50F450(recipient);
	case 2:
		return indirect ? shop_gold_indirect(recipient, amount) : sub_5104F0(recipient, amount);
	default:
		return indirect ? shop_unlink_indirect(session, item) : sub_50E7A0(session, item);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func shopRemovalFixtureCall(entry, route int, recipient, item *server.Object, session *server.TradeSession, amount uint16) uint32 {
	return uint32(C.shop_removal_call(C.int(entry), C.int(route), asObjectC(recipient), asObjectC(item), unsafe.Pointer(session), C.short(amount)))
}
