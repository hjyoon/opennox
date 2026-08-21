// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../break_init_4f0570.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

struct nox_object_t {
	uint32_t xstatus;
};

typedef void (*break_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "xstatus must remain exact uint32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_breakInit_4F0570, break_init_fn: 1, default: 0),
	"BreakInit must use a native object pointer and void callback result");

static nox_object_t* observed_unit;
static uint32_t observed_bit;
static unsigned int set_calls;

static void set_xstatus(nox_object_t* unit, uint32_t bit) {
	observed_unit = unit;
	observed_bit = bit;
	set_calls++;
	unit->xstatus |= bit;
}

void nox_xxx_breakInit_4F0570(nox_object_t* unit) {
	if (((uint8_t)unit->xstatus & UINT8_C(0x0E)) == 0)
		set_xstatus(unit, UINT32_C(2));
}

int main(void) {
	break_init_fn const init = nox_xxx_breakInit_4F0570;
	nox_object_t clear = {.xstatus = UINT32_C(0x0E000001)};
	nox_object_t unrelated = {.xstatus = UINT32_C(0xFFFFFFF1)};
	nox_object_t masked[] = {
		{.xstatus = UINT32_C(2)},
		{.xstatus = UINT32_C(4)},
		{.xstatus = UINT32_C(8)},
		{.xstatus = UINT32_C(0xFFFFFF0E)},
	};
	size_t index;

	init(&clear);
	if (observed_unit != &clear || observed_bit != UINT32_C(2) || set_calls != 1 ||
		clear.xstatus != UINT32_C(0x0E000003))
		return __LINE__;

	init(&unrelated);
	if (observed_unit != &unrelated || observed_bit != UINT32_C(2) || set_calls != 2 ||
		unrelated.xstatus != UINT32_C(0xFFFFFFF3))
		return __LINE__;

	for (index = 0; index < sizeof(masked) / sizeof(masked[0]); ++index) {
		uint32_t const before = masked[index].xstatus;
		init(&masked[index]);
		if (set_calls != 2 || masked[index].xstatus != before)
			return __LINE__;
	}
	return 0;
}
