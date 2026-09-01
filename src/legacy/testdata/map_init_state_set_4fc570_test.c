#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../map_init_state_set_4fc570.h"

typedef int32_t (*map_init_state_set_fn_4fc570)(int32_t);

_Static_assert(CHAR_BIT == 8, "map-init state bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "map-init state must remain one dword");
_Static_assert(
	_Generic(&nox_xxx_resetMapInit_4FC570,
		map_init_state_set_fn_4fc570: 1,
		default: 0),
	"map-init state setter must retain its exact int32 ABI");

static int32_t map_init_state;

int32_t nox_xxx_resetMapInit_4FC570(int32_t value) {
	map_init_state = value;
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
	map_init_state_set_fn_4fc570 const setter = nox_xxx_resetMapInit_4FC570;

	for (unsigned int i = 0; i < sizeof(values) / sizeof(values[0]); ++i) {
		assert(setter(values[i]) == values[i]);
		assert(map_init_state == values[i]);
	}
	return 0;
}
