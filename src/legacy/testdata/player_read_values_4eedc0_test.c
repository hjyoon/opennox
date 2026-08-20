#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_read_values_4eedc0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*player_read_values_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "arguments and result must remain exact int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_plrReadVals_4EEDC0, player_read_values_fn: 1, default: 0),
	"player value initialization must use a native object pointer and exact int32 argument/result");

static nox_object_t* observed_unit;
static int32_t observed_reward_arg;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_plrReadVals_4EEDC0(nox_object_t* unit, int32_t reward_arg) {
	observed_unit = unit;
	observed_reward_arg = reward_arg;
	++observed_calls;
	return next_result;
}

static void check_call(
		player_read_values_fn read_values,
		nox_object_t* unit,
		int32_t reward_arg,
		int32_t result) {
	next_result = result;
	assert(read_values(unit, reward_arg) == result);
	assert(observed_unit == unit);
	assert(observed_reward_arg == reward_arg);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_read_values_fn const read_values = nox_xxx_plrReadVals_4EEDC0;

	check_call(read_values, &unit, INT32_C(0), INT32_C(1));
	check_call(read_values, &unit, INT32_MAX, INT32_MIN);
	check_call(read_values, &unit, INT32_MIN, INT32_MAX);
	check_call(read_values, NULL, INT32_C(-1), INT32_C(0));
	assert(observed_calls == 4);
	assert(observed_unit == NULL);
	return 0;
}
