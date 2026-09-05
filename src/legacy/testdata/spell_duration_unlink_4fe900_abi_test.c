#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_unlink_4fe900.h"

typedef void (*spell_duration_unlink_fn)(void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE900 record must remain pointer-sized");
_Static_assert(
	_Generic(&sub_4FE900, spell_duration_unlink_fn: 1, default: 0),
	"004FE900 must receive one native pointer");

static void* observed_record;
static unsigned int observed_calls;

void sub_4FE900(void* record) {
	observed_record = record;
	++observed_calls;
}

static void check_call(spell_duration_unlink_fn unlink_spell, void* record) {
	unlink_spell(record);
	assert(observed_record == record);
}

int main(void) {
	static unsigned char first_record;
	static unsigned char second_record;
	spell_duration_unlink_fn const unlink_spell = sub_4FE900;

	check_call(unlink_spell, NULL);
	check_call(unlink_spell, &first_record);
	check_call(unlink_spell, &second_record);
	assert(observed_calls == 3);
	return 0;
}
