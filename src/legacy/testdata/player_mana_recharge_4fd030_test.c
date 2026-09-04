#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_mana_recharge_4fd030.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef uint16_t (*player_mana_recharge_fn)(nox_object_t*, int16_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int16_t) == 2, "mana amount and result must remain words");
_Static_assert(sizeof(uint16_t) == 2, "mana amount and result must remain words");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4FD030, player_mana_recharge_fn: 1, default: 0),
	"004FD030 must preserve its native-pointer/int16_t/uint16_t ABI");

static nox_object_t* observed_unit;
static int16_t observed_amount;
static uint16_t next_result;

uint16_t sub_4FD030(nox_object_t* unit, int16_t amount) {
	observed_unit = unit;
	observed_amount = amount;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_mana_recharge_fn const recharge = sub_4FD030;

	next_result = UINT16_MAX;
	assert(recharge(&unit, INT16_MIN) == UINT16_MAX);
	assert(observed_unit == &unit);
	assert(observed_amount == INT16_MIN);

	next_result = UINT16_C(0x8000);
	assert(recharge(NULL, INT16_MAX) == UINT16_C(0x8000));
	assert(observed_unit == NULL);
	assert(observed_amount == INT16_MAX);
	return 0;
}
