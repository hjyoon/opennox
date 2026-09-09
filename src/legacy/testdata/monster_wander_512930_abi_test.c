#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../monster_wander_512930.h"

typedef void (*monster_wander_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_scriptMonsterRoam_512930, monster_wander_fn: 1, default: 0),
	"00512930 must receive one native object pointer");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_xxx_scriptMonsterRoam_512930(nox_object_t* unit) {
	observed_unit = unit;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	monster_wander_fn const wander = nox_xxx_scriptMonsterRoam_512930;

	wander(&unit);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);
	wander(NULL);
	assert(observed_unit == NULL);
	return 0;
}
