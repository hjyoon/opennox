#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../spell_mana_preflight_4fcef0.h"

typedef int32_t (*spell_mana_preflight_fn)(nox_object_t*, int32_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell IDs, count, and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_spellCheckSmth_4FCEF0, spell_mana_preflight_fn: 1, default: 0),
	"004FCEF0 must preserve two native pointers and signed dword count/result");

struct nox_object_t {
	uint32_t marker;
};

static nox_object_t* observed_unit;
static int32_t* observed_sequence;
static int32_t observed_count;

int32_t nox_xxx_spellCheckSmth_4FCEF0(nox_object_t* unit, int32_t* sequence, int32_t count) {
	observed_unit = unit;
	observed_sequence = sequence;
	observed_count = count;
	return INT32_MIN;
}

int main(void) {
	nox_object_t unit = {0};
	int32_t sequence[6] = {74, 75, 114, 115, INT32_MIN, INT32_MAX};
	spell_mana_preflight_fn const preflight = nox_xxx_spellCheckSmth_4FCEF0;

	assert(preflight(&unit, sequence, INT32_MIN) == INT32_MIN);
	assert(observed_unit == &unit);
	assert(observed_sequence == sequence);
	assert(observed_count == INT32_MIN);
	return 0;
}
