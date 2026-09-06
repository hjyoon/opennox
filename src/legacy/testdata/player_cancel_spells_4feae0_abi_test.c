#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_cancel_spells_4feae0.h"

typedef int32_t (*player_cancel_spells_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FEAE0 result must remain 32 bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FEAE0 object argument must remain pointer-sized");
_Static_assert(
	_Generic(&nox_xxx_playerCancelSpells_4FEAE0,
		player_cancel_spells_fn: 1, default: 0),
	"004FEAE0 must preserve its native-pointer and int32 result ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_caster;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_playerCancelSpells_4FEAE0(nox_object_t* caster) {
	observed_caster = caster;
	observed_calls++;
	return next_result;
}

static void check_call(player_cancel_spells_fn cancel, nox_object_t* caster,
	int32_t result) {
	next_result = result;
	assert(cancel(caster) == result);
	assert(observed_caster == caster);
}

int main(void) {
	nox_object_t first = {.marker = UINTPTR_MAX};
	nox_object_t second = {.marker = (uintptr_t)UINT32_C(0x12345678)};
	player_cancel_spells_fn const cancel = nox_xxx_playerCancelSpells_4FEAE0;

	check_call(cancel, NULL, INT32_MIN);
	check_call(cancel, &first, -1);
	check_call(cancel, &second, 0);
	check_call(cancel, &first, INT32_MAX);
	assert(observed_calls == 4);
	assert(first.marker == UINTPTR_MAX);
	assert(second.marker == (uintptr_t)UINT32_C(0x12345678));
	return 0;
}
