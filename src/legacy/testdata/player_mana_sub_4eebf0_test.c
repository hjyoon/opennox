#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_mana_sub_4eebf0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef uintptr_t (*player_mana_sub_fn)(nox_object_t*, int32_t);

_Static_assert(
	_Generic(&nox_xxx_playerManaSub_4EEBF0, player_mana_sub_fn: 1, default: 0),
	"player mana subtraction must retain its native-pointer/int32_t/uintptr_t ABI");

static nox_object_t* observed_unit;
static int32_t observed_amount;
static uintptr_t next_result;

uintptr_t nox_xxx_playerManaSub_4EEBF0(nox_object_t* unit, int32_t amount) {
	observed_unit = unit;
	observed_amount = amount;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_mana_sub_fn const call = nox_xxx_playerManaSub_4EEBF0;

	next_result = UINTPTR_MAX;
	assert(call(&unit, INT32_MIN) == UINTPTR_MAX);
	assert(observed_unit == &unit);
	assert(observed_amount == INT32_MIN);

	next_result = (uintptr_t)0x1234;
	assert(call(NULL, INT32_MAX) == (uintptr_t)0x1234);
	assert(observed_unit == NULL);
	assert(observed_amount == INT32_MAX);
	return 0;
}
