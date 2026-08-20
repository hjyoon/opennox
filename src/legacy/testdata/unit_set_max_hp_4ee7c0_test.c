#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_set_max_hp_4ee7c0.h"

typedef void* (*unit_set_max_hp_callback_t)(nox_object_t*, short);

_Static_assert(CHAR_BIT == 8, "maximum HP bytes must remain eight bits");
_Static_assert(sizeof(short) == 2, "maximum HP argument must remain 16-bit");
_Static_assert(SHRT_MIN == -32768 && SHRT_MAX == 32767, "maximum HP short must be exact signed 16-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitSetMaxHP_4EE7C0, unit_set_max_hp_callback_t: 1, default: 0),
	"UnitSetMaxHP must use a native object pointer, 16-bit C short, and native pointer return");

static nox_object_t* observed_unit;
static short observed_maximum;
static void* returned_health;

void* nox_xxx_unitSetMaxHP_4EE7C0(nox_object_t* unit, short maximum) {
	observed_unit = unit;
	observed_maximum = maximum;
	return returned_health;
}

static void check_word(
	unit_set_max_hp_callback_t set_max_hp,
	nox_object_t* unit,
	void* health,
	short maximum,
	uint16_t bits) {
	returned_health = health;
	assert(set_max_hp(unit, maximum) == health);
	assert(observed_unit == unit);
	assert((uint16_t)observed_maximum == bits);
}

int main(void) {
	uintptr_t unit_storage = UINTPTR_MAX;
	uintptr_t health_storage = UINTPTR_MAX - (uintptr_t)1;
	nox_object_t* const unit = (nox_object_t*)&unit_storage;
	void* const health = &health_storage;
	unit_set_max_hp_callback_t const set_max_hp = nox_xxx_unitSetMaxHP_4EE7C0;

	check_word(set_max_hp, unit, health, 0, UINT16_C(0x0000));
	check_word(set_max_hp, unit, health, SHRT_MAX, UINT16_C(0x7fff));
	check_word(set_max_hp, unit, health, SHRT_MIN, UINT16_C(0x8000));
	check_word(set_max_hp, unit, health, -1, UINT16_C(0xffff));
	assert((uintptr_t)observed_unit == (uintptr_t)unit);
	assert((uintptr_t)returned_health == (uintptr_t)health);

	check_word(set_max_hp, NULL, NULL, -1, UINT16_C(0xffff));
	assert(observed_unit == NULL);
	assert(returned_health == NULL);
	return 0;
}
