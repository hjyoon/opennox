#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_first_4fe930.h"

typedef void* (*spell_duration_first_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE930 result must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_spellCastedFirst_4FE930, spell_duration_first_fn: 1, default: 0),
	"004FE930 must preserve its no-argument native-pointer ABI");

static unsigned char first_record;
static unsigned char second_record;
static unsigned int observed_calls;

void* nox_xxx_spellCastedFirst_4FE930(void) {
	void* const values[] = {NULL, &first_record, &second_record};
	return values[observed_calls++];
}

int main(void) {
	spell_duration_first_fn const first = nox_xxx_spellCastedFirst_4FE930;
	assert(first() == NULL);
	assert(first() == &first_record);
	assert(first() == &second_record);
	assert(observed_calls == 3);
	return 0;
}
