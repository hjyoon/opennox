#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../buff_apply_4ff380.h"

typedef void (*buff_apply_fn)(nox_object_t*, int32_t, int16_t, int8_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FF380 buff must remain a signed dword");
_Static_assert(sizeof(int16_t) == 2, "004FF380 duration must remain a signed word");
_Static_assert(sizeof(int8_t) == 1, "004FF380 power must remain a signed byte");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF380 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_buffApplyTo_4FF380, buff_apply_fn: 1, default: 0),
	"004FF380 must preserve its native-pointer and fixed-width scalar ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_buff;
static int16_t observed_duration;
static int8_t observed_power;
static unsigned int observed_calls;

void nox_xxx_buffApplyTo_4FF380(
	nox_object_t* unit,
	int32_t buff,
	int16_t duration,
	int8_t power
) {
	observed_unit = unit;
	observed_buff = buff;
	observed_duration = duration;
	observed_power = power;
	++observed_calls;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	buff_apply_fn const apply_buff = nox_xxx_buffApplyTo_4FF380;

	apply_buff(&unit, INT32_MIN, INT16_MIN, INT8_MIN);
	assert(observed_unit == &unit);
	assert(observed_buff == INT32_MIN);
	assert(observed_duration == INT16_MIN);
	assert(observed_power == INT8_MIN);
	assert(unit.marker == UINTPTR_MAX);

	apply_buff(NULL, INT32_MAX, INT16_MAX, INT8_MAX);
	assert(observed_unit == NULL);
	assert(observed_buff == INT32_MAX);
	assert(observed_duration == INT16_MAX);
	assert(observed_power == INT8_MAX);
	assert(observed_calls == 2);
	return 0;
}
