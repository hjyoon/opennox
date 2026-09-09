#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../summon_mana_cost_500ca0.h"

typedef int32_t (*summon_mana_cost_fn)(int32_t, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(sub_500CA0(0, NULL)) == 4,
	"00500CA0 call expressions must return exactly four bytes");
_Static_assert(
	_Generic(&sub_500CA0, summon_mana_cost_fn: 1, default: 0),
	"00500CA0 must receive one int32_t and one native object pointer");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell_id;
static nox_object_t* observed_unit;
static int32_t next_result;
static unsigned int observed_calls;

int32_t sub_500CA0(int32_t spell_id, nox_object_t* unit) {
	observed_spell_id = spell_id;
	observed_unit = unit;
	++observed_calls;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	summon_mana_cost_fn const summon_cost = sub_500CA0;

	next_result = INT32_MIN;
	assert(summon_cost(INT32_MIN, &unit) == INT32_MIN);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_unit == &unit);
	assert(observed_unit->marker == UINTPTR_MAX);

	next_result = INT32_MAX;
	assert(summon_cost(INT32_MAX, NULL) == INT32_MAX);
	assert(observed_spell_id == INT32_MAX);
	assert(observed_unit == NULL);
	assert(observed_calls == 2);
	return 0;
}
