package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"

typedef void* (*allocate_shop_t)(void);
typedef void* (*open_shop_t)(nox_object_t*, nox_object_t*);
typedef void (*load_shop_t)(void*);
_Static_assert(_Generic(&nox_xxx_createShopStruct_50E870, allocate_shop_t:1, default:0), "allocate shop ABI");
_Static_assert(_Generic(&nox_xxx_createPlayerShopSession_50E8F0, open_shop_t:1, default:0), "open shop ABI");
_Static_assert(_Generic(&nox_xxx_loadShopItems_50E970, load_shop_t:1, default:0), "load shop ABI");

static allocate_shop_t volatile allocate_shop_indirect = nox_xxx_createShopStruct_50E870;
static open_shop_t volatile open_shop_indirect = nox_xxx_createPlayerShopSession_50E8F0;
static load_shop_t volatile load_shop_indirect = nox_xxx_loadShopItems_50E970;

static void* allocate_shop_fixture(int indirect) {
	return indirect ? allocate_shop_indirect() : nox_xxx_createShopStruct_50E870();
}

static void* open_shop_fixture(int indirect, nox_object_t* player, nox_object_t* merchant) {
	return indirect ? open_shop_indirect(player, merchant) : nox_xxx_createPlayerShopSession_50E8F0(player, merchant);
}

static void load_shop_fixture(int indirect, void* session) {
	if (indirect) {
		load_shop_indirect(session);
	} else {
		nox_xxx_loadShopItems_50E970(session);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func shopConstructorAllocateFixture(route int) *server.TradeSession {
	return (*server.TradeSession)(C.allocate_shop_fixture(C.int(route)))
}

func shopConstructorOpenFixture(route int, player, merchant *server.Object) *server.TradeSession {
	return (*server.TradeSession)(C.open_shop_fixture(C.int(route), asObjectC(player), asObjectC(merchant)))
}

func shopConstructorLoadFixture(route int, session *server.TradeSession) {
	C.load_shop_fixture(C.int(route), unsafe.Pointer(session))
}
