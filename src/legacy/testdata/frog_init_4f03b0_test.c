// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../frog_init_4f03b0.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct frog_init_update_data {
	uint8_t delay;
	uint8_t field_1;
	uint8_t field_2;
} frog_init_update_data;

struct nox_object_t {
	uint16_t direction_1;
	uint16_t direction_2;
	frog_init_update_data* update;
};

typedef int32_t (*frog_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "result must remain exact int32");
_Static_assert(sizeof(uint16_t) == 2, "Direction2 must remain exact uint16");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(frog_init_update_data) == 3, "FrogInit prefix size");
_Static_assert(offsetof(frog_init_update_data, delay) == 0, "delay offset");
_Static_assert(offsetof(frog_init_update_data, field_1) == 1, "field 1 offset");
_Static_assert(offsetof(frog_init_update_data, field_2) == 2, "field 2 offset");
_Static_assert(
	_Generic(&nox_xxx_initFrog_4F03B0, frog_init_fn: 1, default: 0),
	"FrogInit must use a native object pointer and fixed-width int32 result");

static int32_t random_values[2];
static unsigned int random_index;
static nox_object_t* observed_unit;

static int32_t next_random(void) {
	return random_values[random_index++];
}

int32_t nox_xxx_initFrog_4F03B0(nox_object_t* unit) {
	frog_init_update_data* const update = unit->update;
	int32_t result;
	observed_unit = unit;
	update->delay = (uint8_t)next_random();
	update->field_1 = UINT8_C(1);
	update->field_2 = UINT8_C(0);
	result = next_random();
	unit->direction_2 = (uint16_t)result;
	return result;
}

int main(void) {
	frog_init_update_data update = {
		.delay = UINT8_C(0x11),
		.field_1 = UINT8_C(0x22),
		.field_2 = UINT8_C(0x33),
	};
	nox_object_t unit = {
		.direction_1 = UINT16_C(0x1357),
		.direction_2 = UINT16_C(0x2468),
		.update = &update,
	};
	frog_init_fn const init = nox_xxx_initFrog_4F03B0;
	random_values[0] = INT32_C(0x1234563A);
	random_values[1] = INT32_C(0x1234ABCD);

	if (init(&unit) != INT32_C(0x1234ABCD))
		return __LINE__;
	if (observed_unit != &unit || random_index != 2)
		return __LINE__;
	if (update.delay != UINT8_C(0x3A) || update.field_1 != 1 || update.field_2 != 0)
		return __LINE__;
	if (unit.direction_1 != UINT16_C(0x1357) || unit.direction_2 != UINT16_C(0xABCD))
		return __LINE__;
	return 0;
}
