#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_buff_clear_4ff580.h"

typedef void (*unit_buff_clear_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF580 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_unitClearBuffs_4FF580, unit_buff_clear_fn: 1, default: 0),
	"004FF580 must preserve its native-pointer and void-result ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static unsigned int observed_calls;

void nox_xxx_unitClearBuffs_4FF580(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	unit_buff_clear_fn const clear_buffs = nox_xxx_unitClearBuffs_4FF580;

	clear_buffs(&unit);
	assert(observed_unit == &unit);
	assert(unit.marker == UINTPTR_MAX);
	clear_buffs(NULL);
	assert(observed_unit == NULL);
	assert(observed_calls == 2);
	return 0;
}
