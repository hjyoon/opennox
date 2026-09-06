#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_selective_cancel_4feb10.h"

typedef void (*spell_duration_selective_cancel_fn)(int32_t, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FEB10 spell ID must remain exact int32");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FEB10 caster argument must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_spellCancelDurSpell_4FEB10,
		spell_duration_selective_cancel_fn: 1, default: 0),
	"004FEB10 must preserve its exact int32 and native-pointer void ABI");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell_id;
static nox_object_t* observed_caster;
static unsigned int observed_calls;

void nox_xxx_spellCancelDurSpell_4FEB10(int32_t spell_id, nox_object_t* caster) {
	observed_spell_id = spell_id;
	observed_caster = caster;
	++observed_calls;
}

static void check_call(spell_duration_selective_cancel_fn cancel,
	int32_t spell_id, nox_object_t* caster) {
	cancel(spell_id, caster);
	assert(observed_spell_id == spell_id);
	assert(observed_caster == caster);
}

int main(void) {
	nox_object_t first = {.marker = UINTPTR_MAX};
	nox_object_t second = {.marker = (uintptr_t)UINT32_C(0x12345678)};
	spell_duration_selective_cancel_fn const cancel =
		nox_xxx_spellCancelDurSpell_4FEB10;

	check_call(cancel, INT32_MIN, NULL);
	check_call(cancel, INT32_C(-1), &first);
	check_call(cancel, INT32_C(0), &second);
	check_call(cancel, INT32_MAX, &first);
	assert(observed_calls == 4);
	assert(first.marker == UINTPTR_MAX);
	assert(second.marker == (uintptr_t)UINT32_C(0x12345678));
	return 0;
}
