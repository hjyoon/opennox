#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_free_recursive_4fe980.h"

typedef void (*spell_duration_free_recursive_fn)(void*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t), "004FE980 record must remain pointer-sized");
_Static_assert(
	_Generic(&sub_4FE980, spell_duration_free_recursive_fn: 1, default: 0),
	"004FE980 must receive one native pointer");

static void* observed_record;
static unsigned int observed_calls;

void sub_4FE980(void* record) {
	observed_record = record;
	++observed_calls;
}

static void check_call(spell_duration_free_recursive_fn free_recursive, void* record) {
	free_recursive(record);
	assert(observed_record == record);
}

int main(void) {
	static unsigned char first_record;
	static unsigned char second_record;
	spell_duration_free_recursive_fn const free_recursive = sub_4FE980;

	check_call(free_recursive, NULL);
	check_call(free_recursive, &first_record);
	check_call(free_recursive, &second_record);
	assert(observed_calls == 3);
	return 0;
}
