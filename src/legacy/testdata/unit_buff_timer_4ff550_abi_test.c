#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_buff_timer_4ff550.h"

typedef uint32_t (*unit_buff_timer_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FF550 buff must remain a signed dword");
_Static_assert(sizeof(uint16_t) == 2, "004FF550 duration must remain a word");
_Static_assert(sizeof(uint32_t) == 4, "004FF550 result must remain a dword");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF550 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_unitGetBuffTimer_4FF550, unit_buff_timer_fn: 1, default: 0),
	"004FF550 must preserve its native-pointer, signed-dword, and dword-result ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_buff;
static unsigned int observed_calls;

uint32_t nox_xxx_unitGetBuffTimer_4FF550(nox_object_t* unit, int32_t buff) {
	observed_unit = unit;
	observed_buff = buff;
	++observed_calls;
	return buff == INT32_MIN ? (uint32_t)UINT16_MAX : UINT32_C(0x00008000);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_buff_timer_fn const get_timer = nox_xxx_unitGetBuffTimer_4FF550;

	assert(get_timer(&unit, INT32_MIN) == UINT32_C(0x0000ffff));
	assert(observed_unit == &unit);
	assert(observed_buff == INT32_MIN);
	assert(unit.marker == UINTPTR_MAX);
	assert(get_timer(NULL, INT32_MAX) == UINT32_C(0x00008000));
	assert(observed_unit == NULL);
	assert(observed_buff == INT32_MAX);
	assert(observed_calls == 2);
	return 0;
}
