#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../server__gamemech__explevel.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef void (*direct_experience_grant_fn)(nox_object_t*, float);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(float) == 4, "award must remain binary32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_plyrGiveExp_4EF3A0_exp_level, direct_experience_grant_fn: 1, default: 0),
	"direct experience grant must receive one native object pointer and one binary32 award");

static nox_object_t* observed_unit;
static float observed_award;
static unsigned int observed_calls;

void nox_xxx_plyrGiveExp_4EF3A0_exp_level(nox_object_t* unit, float award) {
	observed_unit = unit;
	observed_award = award;
	++observed_calls;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	direct_experience_grant_fn const grant = nox_xxx_plyrGiveExp_4EF3A0_exp_level;

	grant(&unit, -17.25f);
	assert(observed_unit == &unit);
	assert(observed_award == -17.25f);
	grant(NULL, 0.0f);
	assert(observed_calls == 2);
	assert(observed_unit == NULL);
	return 0;
}
