#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_power_4fe7b0.h"

typedef int32_t (*spell_power_fn)(int32_t, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID and power must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_spellGetPower_4FE7B0, spell_power_fn: 1, default: 0),
	"004FE7B0 must preserve its signed-dword and native-pointer ABI");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell;
static nox_object_t* observed_caster;
static int32_t next_result;

int32_t nox_xxx_spellGetPower_4FE7B0(int32_t spell_id, nox_object_t* caster) {
	observed_spell = spell_id;
	observed_caster = caster;
	return next_result;
}

int main(void) {
	nox_object_t caster = {.marker = UINTPTR_MAX};
	spell_power_fn const power = nox_xxx_spellGetPower_4FE7B0;

	next_result = INT32_MIN;
	assert(power(INT32_MAX, &caster) == INT32_MIN);
	assert(observed_spell == INT32_MAX);
	assert(observed_caster == &caster);

	next_result = INT32_MAX;
	assert(power(INT32_MIN, NULL) == INT32_MAX);
	assert(observed_spell == INT32_MIN);
	assert(observed_caster == NULL);
	return 0;
}
