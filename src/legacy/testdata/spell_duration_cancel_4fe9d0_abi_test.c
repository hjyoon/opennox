#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_cancel_4fe9d0.h"

typedef uint8_t (*spell_duration_cancel_fn)(void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint8_t) == 1, "004FE9D0 result must remain one byte");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE9D0 record must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_spellCancelSpellDo_4FE9D0, spell_duration_cancel_fn: 1, default: 0),
	"004FE9D0 must receive one native pointer and return exact uint8_t");

static void* observed_record;
static uint8_t next_result;
static unsigned int observed_calls;

uint8_t nox_xxx_spellCancelSpellDo_4FE9D0(void* record) {
	observed_record = record;
	++observed_calls;
	return next_result;
}

static void check_call(spell_duration_cancel_fn cancel, void* record, uint8_t result) {
	next_result = result;
	assert(cancel(record) == result);
	assert(observed_record == record);
}

int main(void) {
	static unsigned char first_record;
	static unsigned char second_record;
	spell_duration_cancel_fn const cancel = nox_xxx_spellCancelSpellDo_4FE9D0;

	check_call(cancel, NULL, UINT8_C(0));
	check_call(cancel, &first_record, UINT8_C(127));
	check_call(cancel, &second_record, UINT8_C(128));
	check_call(cancel, &first_record, UINT8_MAX);
	assert(observed_calls == 4);
	return 0;
}
