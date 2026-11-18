package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"

typedef void (*stock_decrement_t)(nox_object_t*, void*);
typedef int (*stock_index_t)(void*, nox_object_t*);
typedef int (*stock_match_t)(nox_object_t*, void*);
_Static_assert(_Generic(&sub_510320, stock_decrement_t:1, default:0), "stock decrement ABI");
_Static_assert(_Generic(&nox_xxx_getSomeShopData_5103A0, stock_index_t:1, default:0), "stock index ABI");
_Static_assert(_Generic(&sub_5103F0, stock_match_t:1, default:0), "stock definition match ABI");

static stock_decrement_t volatile stock_decrement_indirect = sub_510320;
static stock_index_t volatile stock_index_indirect = nox_xxx_getSomeShopData_5103A0;
static stock_match_t volatile stock_match_indirect = sub_5103F0;

static void stock_decrement_fixture(int indirect, nox_object_t* item, void* session) {
	if (indirect) stock_decrement_indirect(item, session);
	else sub_510320(item, session);
}

static int stock_index_fixture(int indirect, void* session, nox_object_t* item) {
	return indirect ? stock_index_indirect(session, item) : nox_xxx_getSomeShopData_5103A0(session, item);
}

static int stock_match_fixture(int indirect, nox_object_t* item, void* definition) {
	return indirect ? stock_match_indirect(item, definition) : sub_5103F0(item, definition);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callShopDefinitionDecrement510320(route int, session *server.TradeSession, item *server.Object) {
	C.stock_decrement_fixture(C.int(route), asObjectC(item), unsafe.Pointer(session))
}

func callShopDefinitionIndex5103A0(route int, session *server.TradeSession, item *server.Object) int {
	return int(C.stock_index_fixture(C.int(route), unsafe.Pointer(session), asObjectC(item)))
}

func callShopDefinitionMatch5103F0(route int, item *server.Object, definition *server.ShopkeeperItemDefinition) int {
	return int(C.stock_match_fixture(C.int(route), asObjectC(item), unsafe.Pointer(definition)))
}
