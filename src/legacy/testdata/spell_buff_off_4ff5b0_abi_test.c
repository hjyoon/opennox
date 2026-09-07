#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_buff_off_4ff5b0.h"

typedef int32_t (*spell_buff_off_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"004FF5B0 buff and result must remain signed dwords");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF5B0 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_spellBuffOff_4FF5B0, spell_buff_off_fn: 1, default: 0),
	"004FF5B0 must preserve its native-pointer and signed-dword ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_buff;
static unsigned int observed_calls;

int32_t nox_xxx_spellBuffOff_4FF5B0(nox_object_t* unit, int32_t buff) {
	observed_unit = unit;
	observed_buff = buff;
	++observed_calls;
	return buff;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	spell_buff_off_fn const buff_off = nox_xxx_spellBuffOff_4FF5B0;

	assert(buff_off(&unit, INT32_MIN) == INT32_MIN);
	assert(observed_unit == &unit);
	assert(observed_buff == INT32_MIN);
	assert(unit.marker == UINTPTR_MAX);
	assert(buff_off(NULL, INT32_MAX) == INT32_MAX);
	assert(observed_unit == NULL);
	assert(observed_buff == INT32_MAX);
	assert(observed_calls == 2);
	return 0;
}
