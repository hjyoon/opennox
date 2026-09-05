#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../collision_enchant_4fdf90.h"

typedef void (*collision_enchant_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_collide_4FDF90, collision_enchant_fn: 1, default: 0),
	"004FDF90 must preserve both native object pointers");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_source;
static nox_object_t* observed_target;

void nox_xxx_collide_4FDF90(nox_object_t* source, nox_object_t* target) {
	observed_source = source;
	observed_target = target;
}

int main(void) {
	nox_object_t source = {.marker = UINTPTR_MAX};
	nox_object_t target = {.marker = UINTPTR_MAX - 1};
	collision_enchant_fn const collide = nox_xxx_collide_4FDF90;

	collide(&source, &target);
	assert(observed_source == &source);
	assert(observed_target == &target);
	assert(observed_source->marker == UINTPTR_MAX);
	assert(observed_target->marker == UINTPTR_MAX - 1);

	collide(NULL, NULL);
	assert(observed_source == NULL);
	assert(observed_target == NULL);
	return 0;
}
