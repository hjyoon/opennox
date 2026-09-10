#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_banish_5017f0.h"

typedef void (*unit_banish_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_banishUnit_5017F0, unit_banish_fn: 1, default: 0),
	"005017F0 must receive one native object pointer");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_xxx_banishUnit_5017F0(nox_object_t* unit) {
	observed_unit = unit;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_banish_fn const banish = nox_xxx_banishUnit_5017F0;

	banish(&unit);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);
	banish(NULL);
	assert(observed_unit == NULL);
	return 0;
}
