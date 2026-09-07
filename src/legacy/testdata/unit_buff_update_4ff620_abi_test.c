#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_buff_update_4ff620.h"

typedef void (*unit_buff_update_fn)(nox_object_t*);

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF620 unit must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_updateUnitBuffs_4FF620, unit_buff_update_fn: 1, default: 0),
	"004FF620 must preserve its native-pointer and void-result ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static unsigned int observed_calls;

void nox_xxx_updateUnitBuffs_4FF620(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
}

static void call_stack_object(unit_buff_update_fn update) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	update(&unit);
	assert(observed_unit == &unit);
	assert(unit.marker == UINTPTR_MAX);
}

int main(void) {
	unit_buff_update_fn const update = nox_xxx_updateUnitBuffs_4FF620;
	call_stack_object(update);
	update(NULL);
	assert(observed_unit == NULL);
	assert(observed_calls == 2);
	return 0;
}
