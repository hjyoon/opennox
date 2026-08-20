#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_mana_add_4eeb80.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef uint16_t (*player_mana_add_fn)(nox_object_t*, int16_t);

_Static_assert(
	_Generic(&nox_xxx_playerManaAdd_4EEB80, player_mana_add_fn: 1, default: 0),
	"player mana addition must retain its native-pointer/int16_t/uint16_t ABI");

static nox_object_t* observed_unit;
static int16_t observed_amount;
static uint16_t next_result;

uint16_t nox_xxx_playerManaAdd_4EEB80(nox_object_t* unit, int16_t amount) {
	observed_unit = unit;
	observed_amount = amount;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_mana_add_fn const call = nox_xxx_playerManaAdd_4EEB80;

	next_result = UINT16_MAX;
	assert(call(&unit, INT16_MIN) == UINT16_MAX);
	assert(observed_unit == &unit);
	assert(observed_amount == INT16_MIN);

	next_result = UINT16_C(0x8000);
	assert(call(NULL, INT16_MAX) == UINT16_C(0x8000));
	assert(observed_unit == NULL);
	assert(observed_amount == INT16_MAX);
	return 0;
}
