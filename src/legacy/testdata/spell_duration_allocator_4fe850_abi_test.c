#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../spell_duration_allocator_4fe850.h"

typedef int32_t (*spell_duration_allocator_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "allocator result must remain a signed dword");
_Static_assert(
	_Generic(&nox_xxx_spellCreateDurations_4FE850, spell_duration_allocator_fn: 1, default: 0),
	"004FE850 must preserve its no-argument signed-dword ABI");

static unsigned int observed_calls;

int32_t nox_xxx_spellCreateDurations_4FE850(void) {
	++observed_calls;
	return INT32_MIN;
}

int main(void) {
	spell_duration_allocator_fn const create = nox_xxx_spellCreateDurations_4FE850;
	assert(create() == INT32_MIN);
	assert(observed_calls == 1);
	return 0;
}
