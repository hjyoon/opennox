#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_buff_power_4ff570.h"

typedef uint8_t (*unit_buff_power_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FF570 buff must remain a signed dword");
_Static_assert(sizeof(uint8_t) == 1, "004FF570 result must remain a byte");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF570 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_buffGetPower_4FF570, unit_buff_power_fn: 1, default: 0),
	"004FF570 must preserve its native-pointer, signed-dword, and byte-result ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_buff;
static unsigned int observed_calls;

uint8_t nox_xxx_buffGetPower_4FF570(nox_object_t* unit, int32_t buff) {
	observed_unit = unit;
	observed_buff = buff;
	++observed_calls;
	return buff == INT32_MIN ? UINT8_MAX : UINT8_C(0x80);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_buff_power_fn const get_power = nox_xxx_buffGetPower_4FF570;

	assert(get_power(&unit, INT32_MIN) == UINT8_MAX);
	assert(observed_unit == &unit);
	assert(observed_buff == INT32_MIN);
	assert(unit.marker == UINTPTR_MAX);
	assert(get_power(NULL, INT32_MAX) == UINT8_C(0x80));
	assert(observed_unit == NULL);
	assert(observed_buff == INT32_MAX);
	assert(observed_calls == 2);
	return 0;
}
