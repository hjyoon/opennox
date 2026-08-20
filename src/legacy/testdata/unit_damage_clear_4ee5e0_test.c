#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../unit_damage_clear_4ee5e0.h"

typedef void (*unit_damage_clear_callback_t)(nox_object_t*, int);

_Static_assert(sizeof(int) == 4, "damage amount must remain 32-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitDamageClear_4EE5E0, unit_damage_clear_callback_t: 1, default: 0),
	"UnitDamageClear must use one native object pointer and one 32-bit damage amount");

static nox_object_t* observed_unit;
static int observed_damage;

void nox_xxx_unitDamageClear_4EE5E0(nox_object_t* unit, int damage_amount) {
	observed_unit = unit;
	observed_damage = damage_amount;
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const unit = (nox_object_t*)&storage;

	nox_xxx_unitDamageClear_4EE5E0(unit, INT_MIN);
	assert(observed_unit == unit);
	assert(observed_damage == INT_MIN);

	nox_xxx_unitDamageClear_4EE5E0(unit, INT_MAX);
	assert(observed_unit == unit);
	assert(observed_damage == INT_MAX);
	return 0;
}
