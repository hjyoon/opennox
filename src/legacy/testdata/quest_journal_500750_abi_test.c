#include <assert.h>
#include <stdint.h>

#include "../quest_journal_500540.h"

typedef int32_t (*quest_journal_get_int_fn)(char*);

_Static_assert(sizeof(int32_t) == 4,
	"00500750 result must remain exactly one dword");
_Static_assert(
	_Generic(&sub_500750, quest_journal_get_int_fn: 1, default: 0),
	"00500750 must preserve its mutable C-string, signed-dword ABI");

static uint32_t value_bits;

int32_t sub_500750(char* name) {
	if (name[0] == '\0') {
		return 0;
	}
	return (int32_t)value_bits;
}

int main(void) {
	char missing[] = "";
	char name[] = "War01a:Value";
	quest_journal_get_int_fn const get_integer = sub_500750;

	assert(get_integer(missing) == 0);
	value_bits = UINT32_C(0x80000000);
	assert(get_integer(name) == INT32_MIN);
	value_bits = UINT32_C(0xffffffff);
	assert(get_integer(name) == -1);
	value_bits = UINT32_C(0x7fffffff);
	assert(get_integer(name) == INT32_MAX);
	return 0;
}
