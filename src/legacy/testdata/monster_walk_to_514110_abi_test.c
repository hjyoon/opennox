#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../monster_walk_to_514110.h"

typedef void (*monster_walk_to_fn)(nox_object_t*, float, float);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_monsterWalkTo_514110, monster_walk_to_fn: 1, default: 0),
	"00514110 must receive a native object pointer and two floats");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static float observed_x;
static float observed_y;

void nox_xxx_monsterWalkTo_514110(nox_object_t* unit, float x, float y) {
	observed_unit = unit;
	observed_x = x;
	observed_y = y;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	monster_walk_to_fn const walk = nox_xxx_monsterWalkTo_514110;

	walk(&unit, 8640.0f, 3120.0f);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);
	assert(observed_x == 8640.0f);
	assert(observed_y == 3120.0f);
	return 0;
}
