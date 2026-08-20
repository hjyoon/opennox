#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_get_max_mana_4eecb0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef short (*player_get_max_mana_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "mana bytes must remain eight bits");
_Static_assert(sizeof(short) == 2, "maximum-mana return must remain 16-bit");
_Static_assert(SHRT_MIN == -32768 && SHRT_MAX == 32767, "C short must be exact signed 16-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_playerGetMaxMana_4EECB0, player_get_max_mana_fn: 1, default: 0),
	"maximum-mana getter must use one native object pointer and return a 16-bit C short");

static nox_object_t* observed_unit;
static short next_result;

short nox_xxx_playerGetMaxMana_4EECB0(nox_object_t* unit) {
	observed_unit = unit;
	return next_result;
}

static void check_word(player_get_max_mana_fn get_mana, nox_object_t* unit, short value, uint16_t bits) {
	next_result = value;
	assert((uint16_t)get_mana(unit) == bits);
	assert(observed_unit == unit);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_get_max_mana_fn const get_mana = nox_xxx_playerGetMaxMana_4EECB0;

	check_word(get_mana, &unit, 0, UINT16_C(0x0000));
	check_word(get_mana, &unit, 1000, UINT16_C(0x03e8));
	check_word(get_mana, &unit, SHRT_MAX, UINT16_C(0x7fff));
	check_word(get_mana, &unit, SHRT_MIN, UINT16_C(0x8000));
	check_word(get_mana, &unit, -1, UINT16_C(0xffff));
	assert((uintptr_t)observed_unit == (uintptr_t)&unit);

	check_word(get_mana, NULL, 0, UINT16_C(0x0000));
	assert(observed_unit == NULL);
	return 0;
}
