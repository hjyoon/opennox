#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_new_4fe950.h"

typedef void* (*spell_duration_new_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE950 result must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_newSpellDuration_4FE950, spell_duration_new_fn: 1, default: 0),
	"004FE950 must preserve its no-argument native-pointer ABI");

static unsigned char first_record;
static unsigned char second_record;
static unsigned int observed_calls;

void* nox_xxx_newSpellDuration_4FE950(void) {
	void* const values[] = {NULL, &first_record, &second_record};
	return values[observed_calls++];
}

int main(void) {
	spell_duration_new_fn const allocate = nox_xxx_newSpellDuration_4FE950;
	assert(allocate() == NULL);
	assert(allocate() == &first_record);
	assert(allocate() == &second_record);
	assert(observed_calls == 3);
	return 0;
}
