#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_hunt_5157a0.h"

typedef void (*unit_hunt_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(_Generic(&nox_xxx_unitHunt_5157A0, unit_hunt_fn: 1, default: 0),
	"005157A0 must receive one native object pointer");
_Static_assert(_Generic(&nox_server_unit_hunt_5157a0, unit_hunt_fn: 1, default: 0),
	"the Go bridge must receive one native object pointer");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_server_unit_hunt_5157a0(nox_object_t* unit) {
	observed_unit = unit;
}

void nox_xxx_unitHunt_5157A0(nox_object_t* unit) {
	nox_server_unit_hunt_5157a0(unit);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_hunt_fn const hunt = nox_xxx_unitHunt_5157A0;
	hunt(&unit);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);
	hunt(NULL);
	assert(observed_unit == NULL);
	return 0;
}
