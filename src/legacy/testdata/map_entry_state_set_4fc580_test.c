#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../map_entry_state_set_4fc580.h"

typedef int32_t (*map_entry_state_set_fn_4fc580)(int32_t);

_Static_assert(CHAR_BIT == 8, "map-entry state bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "map-entry state must remain one dword");
_Static_assert(
	_Generic(&sub_4FC580,
		map_entry_state_set_fn_4fc580: 1,
		default: 0),
	"map-entry state setter must retain its exact int32 ABI");

static int32_t map_entry_state;

int32_t sub_4FC580(int32_t value) {
	map_entry_state = value;
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
	map_entry_state_set_fn_4fc580 const setter = sub_4FC580;

	for (unsigned int i = 0; i < sizeof(values) / sizeof(values[0]); ++i) {
		assert(setter(values[i]) == values[i]);
		assert(map_entry_state == values[i]);
	}
	return 0;
}
