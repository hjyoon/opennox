// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../direction_init_4f0490.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct direction_init_data {
	int32_t x;
	int32_t y;
} direction_init_data;

struct nox_object_t {
	uint32_t guard_before;
	uint16_t direction_1;
	uint16_t direction_2;
	uint32_t guard_after;
	direction_init_data* init_data;
};

typedef int32_t (*direction_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "direction components and result must remain exact dwords");
_Static_assert(sizeof(uint16_t) == 2, "stored directions must remain exact words");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(direction_init_data) == 8, "direction init-data size");
_Static_assert(offsetof(direction_init_data, x) == 0, "direction X offset");
_Static_assert(offsetof(direction_init_data, y) == 4, "direction Y offset");
_Static_assert(
	_Generic(&sub_4F0490, direction_init_fn: 1, default: 0),
	"DirectionInit must use one native object pointer and return an exact int32_t");

int32_t sub_4F0490(nox_object_t* unit) {
	static uint32_t const table[9] = {
		UINT32_C(160), UINT32_C(192), UINT32_C(224),
		UINT32_C(128), UINT32_C(0), UINT32_C(0),
		UINT32_C(96), UINT32_C(64), UINT32_C(32),
	};
	direction_init_data* const init_data = unit->init_data;
	int32_t const index = init_data->x + INT32_C(3) * init_data->y;
	uint32_t const angle = table[index + INT32_C(4)];

	unit->direction_2 = (uint16_t)angle;
	unit->direction_1 = (uint16_t)angle;
	return (int32_t)angle;
}

int main(void) {
	static uint32_t const expected[3][3] = {
		{UINT32_C(160), UINT32_C(192), UINT32_C(224)},
		{UINT32_C(128), UINT32_C(0), UINT32_C(0)},
		{UINT32_C(96), UINT32_C(64), UINT32_C(32)},
	};
	direction_init_fn const init = sub_4F0490;
	int32_t y;

	for (y = -INT32_C(1); y <= INT32_C(1); ++y) {
		int32_t x;
		for (x = -INT32_C(1); x <= INT32_C(1); ++x) {
			direction_init_data init_data = {.x = x, .y = y};
			nox_object_t unit = {
				.guard_before = UINT32_C(0x11111111),
				.direction_1 = UINT16_C(0xAAAA),
				.direction_2 = UINT16_C(0xBBBB),
				.guard_after = UINT32_C(0x22222222),
				.init_data = &init_data,
			};
			uint32_t const want = expected[y + INT32_C(1)][x + INT32_C(1)];

			if ((uint32_t)init(&unit) != want)
				return __LINE__;
			if (unit.direction_1 != (uint16_t)want || unit.direction_2 != (uint16_t)want)
				return __LINE__;
			if (unit.guard_before != UINT32_C(0x11111111) ||
				unit.guard_after != UINT32_C(0x22222222) || unit.init_data != &init_data)
				return __LINE__;
		}
	}
	return 0;
}
