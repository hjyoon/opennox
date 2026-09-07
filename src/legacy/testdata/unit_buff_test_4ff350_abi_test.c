#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_buff_test_4ff350.h"

typedef int32_t (*unit_buff_test_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FF350 buff must remain a signed dword");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF350 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_testUnitBuffs_4FF350, unit_buff_test_fn: 1, default: 0),
	"004FF350 must preserve its native-pointer and signed-dword ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_buff;
static unsigned int observed_calls;

int32_t nox_xxx_testUnitBuffs_4FF350(nox_object_t* unit, int32_t buff) {
	observed_unit = unit;
	observed_buff = buff;
	++observed_calls;
	return buff == INT32_MIN;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_buff_test_fn const test_buff = nox_xxx_testUnitBuffs_4FF350;

	assert(test_buff(&unit, INT32_MIN) == 1);
	assert(observed_unit == &unit);
	assert(observed_buff == INT32_MIN);
	assert(unit.marker == UINTPTR_MAX);
	assert(test_buff(NULL, INT32_MAX) == 0);
	assert(observed_unit == NULL);
	assert(observed_buff == INT32_MAX);
	assert(observed_calls == 2);
	return 0;
}
