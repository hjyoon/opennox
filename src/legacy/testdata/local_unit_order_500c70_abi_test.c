#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../local_unit_order_500c70.h"

typedef int32_t (*local_unit_order_fn)(int32_t, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "00500C70 scalars must remain four bytes");
_Static_assert(sizeof(nox_xxx_orderUnitLocal_500C70(0, 0)) == 4,
	"00500C70 call expressions must return exactly four bytes");
_Static_assert(
	_Generic(&nox_xxx_orderUnitLocal_500C70,
		local_unit_order_fn: 1,
		default: 0),
	"00500C70 must preserve its two-int32_t ABI");

static int32_t observed_owner;
static int32_t observed_order;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_orderUnitLocal_500C70(int32_t owner, int32_t order_type) {
	observed_owner = owner;
	observed_order = order_type;
	++observed_calls;
	return next_result;
}

static int32_t check_call(
	local_unit_order_fn order_unit,
	int32_t owner,
	int32_t order_type) {
	return order_unit(owner, order_type);
}

int main(void) {
	local_unit_order_fn const order_unit = nox_xxx_orderUnitLocal_500C70;

	next_result = INT32_MIN;
	assert(check_call(order_unit, INT32_MIN, INT32_C(-1985229329)) == INT32_MIN);
	assert(observed_owner == INT32_MIN);
	assert((uint32_t)observed_order == UINT32_C(0x89abcdef));

	next_result = INT32_MAX;
	assert(check_call(order_unit, INT32_MAX, INT32_MAX) == INT32_MAX);
	assert(observed_owner == INT32_MAX);
	assert(observed_order == INT32_MAX);

	next_result = INT32_C(-1);
	assert(check_call(order_unit, INT32_C(-1), INT32_C(-1)) == INT32_C(-1));
	assert(observed_owner == INT32_C(-1));
	assert((uint32_t)observed_order == UINT32_MAX);
	assert(observed_calls == 3);
	return 0;
}
