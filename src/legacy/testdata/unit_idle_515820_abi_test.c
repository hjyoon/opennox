#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_idle_515820.h"

typedef void (*unit_idle_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitIdle_515820, unit_idle_fn: 1, default: 0),
	"00515820 must receive one native object pointer");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_xxx_unitIdle_515820(nox_object_t* unit) {
	observed_unit = unit;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_idle_fn const idle = nox_xxx_unitIdle_515820;

	idle(&unit);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);
	idle(NULL);
	assert(observed_unit == NULL);
	return 0;
}
