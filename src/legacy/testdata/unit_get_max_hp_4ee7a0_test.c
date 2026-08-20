#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_get_max_hp_4ee7a0.h"

typedef short (*unit_get_max_hp_callback_t)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "maximum HP bytes must remain eight bits");
_Static_assert(sizeof(short) == 2, "maximum HP return must remain 16-bit");
_Static_assert(SHRT_MIN == -32768 && SHRT_MAX == 32767, "maximum HP short must be exact signed 16-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitGetMaxHP_4EE7A0, unit_get_max_hp_callback_t: 1, default: 0),
	"UnitGetMaxHP must use one native object pointer and return a 16-bit C short");

static nox_object_t* observed_unit;
static short returned_hp;

short nox_xxx_unitGetMaxHP_4EE7A0(nox_object_t* unit) {
	observed_unit = unit;
	return returned_hp;
}

static void check_word(unit_get_max_hp_callback_t get_hp, nox_object_t* unit, short value, uint16_t bits) {
	returned_hp = value;
	assert((uint16_t)get_hp(unit) == bits);
	assert(observed_unit == unit);
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const unit = (nox_object_t*)&storage;
	unit_get_max_hp_callback_t const get_hp = nox_xxx_unitGetMaxHP_4EE7A0;

	check_word(get_hp, unit, 0, UINT16_C(0x0000));
	check_word(get_hp, unit, SHRT_MAX, UINT16_C(0x7fff));
	check_word(get_hp, unit, SHRT_MIN, UINT16_C(0x8000));
	check_word(get_hp, unit, -1, UINT16_C(0xffff));
	assert((uintptr_t)observed_unit == (uintptr_t)unit);

	check_word(get_hp, NULL, 0, UINT16_C(0x0000));
	assert(observed_unit == NULL);
	return 0;
}
