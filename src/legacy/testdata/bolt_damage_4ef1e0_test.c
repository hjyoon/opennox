#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../bolt_damage_4ef1e0.h"

struct nox_modifier_t {
	uintptr_t marker;
};

typedef double (*bolt_damage_fn)(int32_t, nox_modifier_t*);
typedef double (*bolt_damage_values_fn)(int32_t, uint32_t, uint16_t, float, uint16_t);
typedef uint32_t (*modifier_type_fn)(nox_modifier_t*);
typedef uint16_t (*modifier_minimum_fn)(nox_modifier_t*);
typedef float (*modifier_range_fn)(nox_modifier_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "strength must remain exact signed int32");
_Static_assert(sizeof(uint32_t) == 4, "type index must remain exact uint32");
_Static_assert(sizeof(uint16_t) == 2, "modifier words must remain exact uint16");
_Static_assert(sizeof(float) == 4, "coefficient and range must remain binary32");
_Static_assert(sizeof(double) == 8, "damage result must remain binary64");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_calcBoltDamage_4EF1E0, bolt_damage_fn: 1, default: 0),
	"bolt damage must receive a native modifier pointer");
_Static_assert(
	_Generic(&nox_xxx_calcBoltDamageValues_4EF1E0, bolt_damage_values_fn: 1, default: 0),
	"synthetic bolt damage must use fixed-width scalar fields");
_Static_assert(
	_Generic(&nox_xxx_boltDamageModifierType_4EF1E0, modifier_type_fn: 1, default: 0),
	"modifier type accessor has the wrong ABI");
_Static_assert(
	_Generic(&nox_xxx_boltDamageModifierMinimum_4EF1E0, modifier_minimum_fn: 1, default: 0),
	"modifier minimum accessor has the wrong ABI");
_Static_assert(
	_Generic(&nox_xxx_boltDamageModifierRange_4EF1E0, modifier_range_fn: 1, default: 0),
	"modifier range accessor has the wrong ABI");

static nox_modifier_t* observed_modifier;
static int32_t observed_strength;
static uint32_t observed_type;
static uint16_t observed_required;
static float observed_coefficient;
static uint16_t observed_minimum;

double nox_xxx_calcBoltDamage_4EF1E0(int32_t strength, nox_modifier_t* modifier) {
	observed_strength = strength;
	observed_modifier = modifier;
	return 123.5;
}

double nox_xxx_calcBoltDamageValues_4EF1E0(
		int32_t strength,
		uint32_t type_index,
		uint16_t required_strength,
		float coefficient,
		uint16_t minimum) {
	observed_strength = strength;
	observed_type = type_index;
	observed_required = required_strength;
	observed_coefficient = coefficient;
	observed_minimum = minimum;
	return -45.25;
}

uint32_t nox_xxx_boltDamageModifierType_4EF1E0(nox_modifier_t* modifier) {
	observed_modifier = modifier;
	return UINT32_C(0xfedcba98);
}

uint16_t nox_xxx_boltDamageModifierMinimum_4EF1E0(nox_modifier_t* modifier) {
	observed_modifier = modifier;
	return UINT16_C(0xabcd);
}

float nox_xxx_boltDamageModifierRange_4EF1E0(nox_modifier_t* modifier) {
	observed_modifier = modifier;
	return 17.25f;
}

int main(void) {
	nox_modifier_t modifier = {.marker = UINTPTR_MAX};
	bolt_damage_fn const calculate = nox_xxx_calcBoltDamage_4EF1E0;
	bolt_damage_values_fn const calculate_values = nox_xxx_calcBoltDamageValues_4EF1E0;

	assert(calculate(INT32_MIN, &modifier) == 123.5);
	assert(observed_strength == INT32_MIN);
	assert(observed_modifier == &modifier);

	assert(calculate_values(INT32_MAX, UINT32_MAX, UINT16_MAX, -0.75f, UINT16_C(0x8000)) == -45.25);
	assert(observed_strength == INT32_MAX);
	assert(observed_type == UINT32_MAX);
	assert(observed_required == UINT16_MAX);
	assert(observed_coefficient == -0.75f);
	assert(observed_minimum == UINT16_C(0x8000));

	assert(nox_xxx_boltDamageModifierType_4EF1E0(&modifier) == UINT32_C(0xfedcba98));
	assert(observed_modifier == &modifier);
	assert(nox_xxx_boltDamageModifierMinimum_4EF1E0(&modifier) == UINT16_C(0xabcd));
	assert(observed_modifier == &modifier);
	assert(nox_xxx_boltDamageModifierRange_4EF1E0(&modifier) == 17.25f);
	assert(observed_modifier == &modifier);
	return 0;
}
