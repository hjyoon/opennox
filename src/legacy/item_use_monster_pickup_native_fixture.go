package legacy

/*
#include "GAME4_3.h"

typedef int (*item_use_entry_t)(nox_object_t*, nox_object_t*);
typedef char (*monster_pickup_entry_t)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_useByNetCode_53F8E0, item_use_entry_t:1, default:0), "item use entry ABI");
_Static_assert(_Generic(&nox_xxx_mobActionPickupObject_544B90, monster_pickup_entry_t:1, default:0), "monster pickup entry ABI");

static item_use_entry_t volatile item_use_entry_indirect = nox_xxx_useByNetCode_53F8E0;
static monster_pickup_entry_t volatile monster_pickup_entry_indirect = nox_xxx_mobActionPickupObject_544B90;

typedef struct item_use_trace_t {
	nox_object_t* owner;
	nox_object_t* item;
	unsigned int calls;
} item_use_trace_t;

static item_use_trace_t* item_use_trace;
static int item_use_result_word;

static int item_use_fixture_callback(nox_object_t* owner, nox_object_t* item) {
	if (item_use_trace) {
		item_use_trace->owner = owner;
		item_use_trace->item = item;
		++item_use_trace->calls;
	}
	return item_use_result_word;
}

static void* item_use_fixture_callback_address(void) { return item_use_fixture_callback; }

static int item_use_entry_fixture(int indirect, nox_object_t* owner, nox_object_t* item,
	                             int callback_result, item_use_trace_t* observed) {
	item_use_result_word = callback_result;
	item_use_trace = observed;
	int result = indirect ? item_use_entry_indirect(owner, item) : nox_xxx_useByNetCode_53F8E0(owner, item);
	item_use_trace = NULL;
	return result;
}

static char monster_pickup_entry_fixture(int indirect, nox_object_t* unit) {
	return indirect ? monster_pickup_entry_indirect(unit) : nox_xxx_mobActionPickupObject_544B90(unit);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func itemUseFixtureCallback53F8E0() unsafe.Pointer {
	return C.item_use_fixture_callback_address()
}

func callItemUseNative53F8E0(route int, owner, item *server.Object, raw uint32) (uint32, int, *server.Object, *server.Object) {
	var trace C.item_use_trace_t
	result := C.item_use_entry_fixture(C.int(route), asObjectC(owner), asObjectC(item), C.int(int32(raw)), &trace)
	return uint32(result), int(trace.calls), asObjectS(trace.owner), asObjectS(trace.item)
}

func callMonsterPickupNative544B90(route int, unit *server.Object) int8 {
	return int8(C.monster_pickup_entry_fixture(C.int(route), asObjectC(unit)))
}
