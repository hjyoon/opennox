#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_team_4f6d20.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*team_xfer_fn_4F6D20)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"TeamXfer result must remain int32");
_Static_assert(
	_Generic(&nox_xxx_XFerTeam_4F6D20,
		team_xfer_fn_4F6D20: 1, default: 0),
	"TeamXfer must preserve one native pointer and exact int32 result");

static nox_object_t* observed_object;

int32_t nox_xxx_XFerTeam_4F6D20(nox_object_t* object) {
	observed_object = object;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	team_xfer_fn_4F6D20 const transfer = nox_xxx_XFerTeam_4F6D20;

	assert(transfer(&object) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);

	assert(transfer(NULL) == 0);
	assert(observed_object == NULL);
	return 0;
}
