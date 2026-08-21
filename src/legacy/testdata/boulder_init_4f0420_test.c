// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../boulder_init_4f0420.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

struct nox_object_t {
	uint32_t source_x;
	uint32_t source_y;
	uint32_t target_x;
	uint32_t target_y;
	uint32_t guard;
};

typedef nox_object_t* (*boulder_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "coordinate payloads must remain exact dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitBoulderInit_4F0420, boulder_init_fn: 1, default: 0),
	"BoulderInit must use and return one native object pointer");

static nox_object_t* observed_unit;

nox_object_t* nox_xxx_unitBoulderInit_4F0420(nox_object_t* unit) {
	uint32_t const source_x = unit->source_x;
	uint32_t const source_y = unit->source_y;

	observed_unit = unit;
	unit->target_x = source_x;
	unit->target_y = source_y;
	return unit;
}

int main(void) {
	nox_object_t unit = {
		.source_x = UINT32_C(0x7FA12345),
		.source_y = UINT32_C(0x80000000),
		.target_x = UINT32_C(0x11111111),
		.target_y = UINT32_C(0x22222222),
		.guard = UINT32_C(0xA5A5A5A5),
	};
	boulder_init_fn const init = nox_xxx_unitBoulderInit_4F0420;
	nox_object_t* const result = init(&unit);

	if (observed_unit != &unit || result != &unit)
		return __LINE__;
	if (unit.target_x != UINT32_C(0x7FA12345) ||
		unit.target_y != UINT32_C(0x80000000))
		return __LINE__;
	if (unit.source_x != UINT32_C(0x7FA12345) ||
		unit.source_y != UINT32_C(0x80000000) ||
		unit.guard != UINT32_C(0xA5A5A5A5))
		return __LINE__;
	return 0;
}
