package legacy

/*
#include "server__system__trade.h"

typedef int (*shop_item_type_callback)(nox_object_t*);
_Static_assert(_Generic(&sub_510540, shop_item_type_callback:1, default:0), "quest shop removal predicate ABI");
_Static_assert(_Generic(&sub_5105D0, shop_item_type_callback:1, default:0), "gem predicate ABI");

static shop_item_type_callback volatile quest_item_type = sub_510540;
static shop_item_type_callback volatile gem_item_type = sub_5105D0;

static int shop_item_type_call(nox_object_t* item, int gem, int indirect) {
	if (gem) {
		return indirect ? gem_item_type(item) : sub_5105D0(item);
	}
	return indirect ? quest_item_type(item) : sub_510540(item);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func shopItemTypeFixtureCall(item *server.Object, gem, route int) int {
	return int(C.shop_item_type_call(asObjectC(item), C.int(gem), C.int(route)))
}
