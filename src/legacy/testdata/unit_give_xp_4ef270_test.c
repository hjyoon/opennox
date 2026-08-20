#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../server__object__health.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef double (*unit_give_xp_fn)(nox_object_t*, float);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(float) == 4, "target must remain binary32");
_Static_assert(sizeof(double) == 8, "award must remain binary64");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitGiveXP_4EF270, unit_give_xp_fn: 1, default: 0),
	"UnitGiveXP must receive one native object pointer and one binary32 target");

static nox_object_t* observed_unit;
static float observed_target;

double nox_xxx_unitGiveXP_4EF270(nox_object_t* unit, float target) {
	observed_unit = unit;
	observed_target = target;
	return 123.5;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_give_xp_fn const give_xp = nox_xxx_unitGiveXP_4EF270;
	assert(give_xp(&unit, -17.25f) == 123.5);
	assert(observed_unit == &unit);
	assert(observed_target == -17.25f);
	return 0;
}
