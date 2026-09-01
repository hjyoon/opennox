#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../coop_ability_state_set_4fc670.h"

typedef int32_t (*coop_ability_state_set_fn_4fc670)(int32_t);

_Static_assert(CHAR_BIT == 8, "cooperative-ability state bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "cooperative-ability state must remain one dword");
_Static_assert(
	_Generic(&sub_4FC670,
		coop_ability_state_set_fn_4fc670: 1,
		default: 0),
	"cooperative-ability state setter must retain its exact int32 ABI");

static int32_t coop_ability_state;

int32_t sub_4FC670(int32_t value) {
	coop_ability_state = value;
	return value;
}

int main(void) {
	static const int32_t values[] = {
		INT32_MIN,
		-1,
		0,
		1,
		INT32_MAX,
		(int32_t)UINT32_C(0x89abcdef),
	};
	coop_ability_state_set_fn_4fc670 const setter = sub_4FC670;

	for (unsigned int i = 0; i < sizeof(values) / sizeof(values[0]); ++i) {
		assert(setter(values[i]) == values[i]);
		assert(coop_ability_state == values[i]);
	}
	return 0;
}
