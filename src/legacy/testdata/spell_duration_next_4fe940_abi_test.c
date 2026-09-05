#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_next_4fe940.h"

typedef void* (*spell_duration_next_fn)(void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE940 pointers must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_spellCastedNext_4FE940, spell_duration_next_fn: 1, default: 0),
	"004FE940 must preserve its one-pointer argument and pointer-result ABI");

static unsigned char first_record;
static unsigned char second_record;
static void* observed_record;
static unsigned int observed_calls;

void* nox_xxx_spellCastedNext_4FE940(void* record) {
	observed_record = record;
	++observed_calls;
	if (record == &first_record) {
		return &second_record;
	}
	if (record == &second_record) {
		return NULL;
	}
	return &first_record;
}

int main(void) {
	spell_duration_next_fn const next = nox_xxx_spellCastedNext_4FE940;

	assert(next(NULL) == &first_record);
	assert(observed_record == NULL);
	assert(next(&first_record) == &second_record);
	assert(observed_record == &first_record);
	assert(next(&second_record) == NULL);
	assert(observed_record == &second_record);
	assert(observed_calls == 3);
	return 0;
}
