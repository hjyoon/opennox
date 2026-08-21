// Keep this fixture independent from the Win32-only aggregate legacy headers
// so the retained ABI can be compiled by every supported target frontend.
#include "../player_unit_init_4efe80.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef uint8_t (*player_unit_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint8_t) == 1, "result must remain exact uint8");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitInitPlayer_4EFE80, player_unit_init_fn: 1, default: 0),
	"player initialization must use a native object pointer and uint8 result");

static nox_object_t* observed_unit;
static uint8_t next_result;
static unsigned int observed_calls;

uint8_t nox_xxx_unitInitPlayer_4EFE80(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
	return next_result;
}

static int check_call(
		player_unit_init_fn initialize,
		nox_object_t* unit,
		uint8_t result) {
	next_result = result;
	if (initialize(unit) != result)
		return __LINE__;
	if (observed_unit != unit)
		return __LINE__;
	return 0;
}

int main(void) {
	unsigned char unit_storage = 0;
	nox_object_t* const unit = (nox_object_t*)(void*)&unit_storage;
	player_unit_init_fn const initialize = nox_xxx_unitInitPlayer_4EFE80;
	int line;

	line = check_call(initialize, unit, UINT8_C(0));
	if (line != 0)
		return line;
	line = check_call(initialize, unit, UINT8_C(127));
	if (line != 0)
		return line;
	line = check_call(initialize, NULL, UINT8_C(128));
	if (line != 0)
		return line;
	line = check_call(initialize, NULL, UINT8_MAX);
	if (line != 0)
		return line;
	if (observed_calls != 4)
		return __LINE__;
	return 0;
}
