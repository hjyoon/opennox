#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_set_max_mana_4eecd0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef uintptr_t (*player_set_max_mana_fn)(nox_object_t*, short);

_Static_assert(CHAR_BIT == 8, "mana bytes must remain eight bits");
_Static_assert(sizeof(short) == 2, "maximum-mana argument must remain 16-bit");
_Static_assert(SHRT_MIN == -32768 && SHRT_MAX == 32767, "C short must be exact signed 16-bit");
_Static_assert(sizeof(uintptr_t) == sizeof(void*), "return register must preserve native pointers");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_playerSetMaxMana_4EECD0, player_set_max_mana_fn: 1, default: 0),
	"maximum-mana setter must use a native object pointer, 16-bit C short, and native return register");

static nox_object_t* observed_unit;
static short observed_maximum;
static uintptr_t next_result;

uintptr_t nox_xxx_playerSetMaxMana_4EECD0(nox_object_t* unit, short maximum) {
	observed_unit = unit;
	observed_maximum = maximum;
	return next_result;
}

static void check_word(player_set_max_mana_fn set_mana, nox_object_t* unit, short value, uint16_t bits) {
	next_result = (uintptr_t)unit;
	assert(set_mana(unit, value) == next_result);
	assert(observed_unit == unit);
	assert((uint16_t)observed_maximum == bits);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_set_max_mana_fn const set_mana = nox_xxx_playerSetMaxMana_4EECD0;

	check_word(set_mana, &unit, 0, UINT16_C(0x0000));
	check_word(set_mana, &unit, 1000, UINT16_C(0x03e8));
	check_word(set_mana, &unit, SHRT_MAX, UINT16_C(0x7fff));
	check_word(set_mana, &unit, SHRT_MIN, UINT16_C(0x8000));
	check_word(set_mana, &unit, -1, UINT16_C(0xffff));
	assert((uintptr_t)observed_unit == (uintptr_t)&unit);

	check_word(set_mana, NULL, -1, UINT16_C(0xffff));
	assert(observed_unit == NULL);
	return 0;
}
