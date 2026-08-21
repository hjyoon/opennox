// Keep this fixture independent from the Win32-only aggregate legacy headers
// so the retained ABI can be compiled by every supported target frontend.
#include "../player_reset_4eff10.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*player_reset_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "result must remain exact int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EFF10, player_reset_fn: 1, default: 0),
	"player reset must use a native object pointer and int32 result");

static nox_object_t* observed_unit;
static int32_t next_result;
static unsigned int observed_calls;

int32_t sub_4EFF10(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
	return next_result;
}

static int check_call(
		player_reset_fn reset,
		nox_object_t* unit,
		int32_t result) {
	next_result = result;
	if (reset(unit) != result)
		return __LINE__;
	if (observed_unit != unit)
		return __LINE__;
	return 0;
}

int main(void) {
	unsigned char unit_storage = 0;
	nox_object_t* const unit = (nox_object_t*)(void*)&unit_storage;
	player_reset_fn const reset = sub_4EFF10;
	int line;

	line = check_call(reset, unit, INT32_MIN);
	if (line != 0)
		return line;
	line = check_call(reset, unit, INT32_MAX);
	if (line != 0)
		return line;
	line = check_call(reset, NULL, INT32_C(0));
	if (line != 0)
		return line;
	line = check_call(reset, NULL, -INT32_C(559023410));
	if (line != 0)
		return line;
	if (observed_calls != 4)
		return __LINE__;
	return 0;
}
