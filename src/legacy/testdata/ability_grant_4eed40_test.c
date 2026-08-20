#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../ability_grant_4eed40.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef void (*ability_grant_fn)(nox_object_t*, int8_t, int32_t);

_Static_assert(CHAR_BIT == 8, "count byte must remain eight bits");
_Static_assert(sizeof(int8_t) == 1, "count must remain an exact signed byte");
_Static_assert(INT8_MIN == -128 && INT8_MAX == 127, "count must preserve signed-byte limits");
_Static_assert(sizeof(int32_t) == 4, "reward argument must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_abilGivePlayerAll_4EED40, ability_grant_fn: 1, default: 0),
	"ability grant must use a native object pointer, signed-byte count, and signed int32 reward argument");

static nox_object_t* observed_unit;
static int8_t observed_count;
static int32_t observed_reward_arg;
static unsigned int observed_calls;

void nox_xxx_abilGivePlayerAll_4EED40(
		nox_object_t* unit, int8_t count, int32_t reward_arg) {
	observed_unit = unit;
	observed_count = count;
	observed_reward_arg = reward_arg;
	++observed_calls;
}

static void check_call(
		ability_grant_fn grant,
		nox_object_t* unit,
		int8_t count,
		int32_t reward_arg) {
	grant(unit, count, reward_arg);
	assert(observed_unit == unit);
	assert(observed_count == count);
	assert(observed_reward_arg == reward_arg);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	ability_grant_fn const grant = nox_xxx_abilGivePlayerAll_4EED40;

	check_call(grant, &unit, INT8_C(10), INT32_C(0));
	check_call(grant, &unit, INT8_MAX, INT32_MAX);
	check_call(grant, &unit, INT8_MIN, INT32_MIN);
	check_call(grant, NULL, INT8_C(-1), INT32_C(-1));
	assert(observed_calls == 4);
	assert((uintptr_t)observed_unit == (uintptr_t)NULL);
	return 0;
}
