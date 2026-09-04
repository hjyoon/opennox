#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_precheck_4fd0e0.h"

typedef int32_t (*spell_precheck_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4FD0E0, spell_precheck_fn: 1, default: 0),
	"004FD0E0 must preserve its native-pointer and signed-dword ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_spell;
static int32_t next_result;

int32_t sub_4FD0E0(nox_object_t* unit, int32_t spell_id) {
	observed_unit = unit;
	observed_spell = spell_id;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	spell_precheck_fn const precheck = sub_4FD0E0;

	next_result = INT32_MIN;
	assert(precheck(&unit, INT32_MAX) == INT32_MIN);
	assert(observed_unit == &unit);
	assert(observed_spell == INT32_MAX);

	next_result = INT32_MAX;
	assert(precheck(NULL, INT32_MIN) == INT32_MAX);
	assert(observed_unit == NULL);
	assert(observed_spell == INT32_MIN);
	return 0;
}
