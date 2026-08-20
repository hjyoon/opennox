#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../server__gamemech__explevel.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef void (*experience_level_update_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "level bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EF2E0_exp_level, experience_level_update_fn: 1, default: 0),
	"experience-level update must receive one native object pointer");

static nox_object_t* observed_unit;
static unsigned int observed_calls;

void sub_4EF2E0_exp_level(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
}

static void check_call(experience_level_update_fn update, nox_object_t* unit) {
	update(unit);
	assert(observed_unit == unit);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	experience_level_update_fn const update = sub_4EF2E0_exp_level;

	check_call(update, &unit);
	check_call(update, NULL);
	assert(observed_calls == 2);
	assert(observed_unit == NULL);
	return 0;
}
