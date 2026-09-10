#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../charm_lifecycle_5011f0.h"

typedef int32_t (*charm_phase_fn)(nox_dur_spell_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"charm phase results must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_charmCreature1_5011F0, charm_phase_fn: 1, default: 0),
	"005011F0 must receive one native duration-record pointer");
_Static_assert(
	_Generic(&nox_xxx_charmCreatureFinish_5013E0, charm_phase_fn: 1, default: 0),
	"005013E0 must receive one native duration-record pointer");
_Static_assert(
	_Generic(&nox_xxx_charmCreature2_501690, charm_phase_fn: 1, default: 0),
	"00501690 must receive one native duration-record pointer");

struct nox_dur_spell_t {
	uintptr_t marker;
};

static nox_dur_spell_t* observed_start;
static nox_dur_spell_t* observed_finish;
static nox_dur_spell_t* observed_cancel;

int32_t nox_xxx_charmCreature1_5011F0(nox_dur_spell_t* record) {
	observed_start = record;
	return INT32_MIN;
}

int32_t nox_xxx_charmCreatureFinish_5013E0(nox_dur_spell_t* record) {
	observed_finish = record;
	return INT32_MAX;
}

int32_t nox_xxx_charmCreature2_501690(nox_dur_spell_t* record) {
	observed_cancel = record;
	return -1;
}

int main(void) {
	nox_dur_spell_t record = {.marker = UINTPTR_MAX};
	charm_phase_fn const start = nox_xxx_charmCreature1_5011F0;
	charm_phase_fn const finish = nox_xxx_charmCreatureFinish_5013E0;
	charm_phase_fn const cancel = nox_xxx_charmCreature2_501690;

	assert(start(&record) == INT32_MIN);
	assert(finish(&record) == INT32_MAX);
	assert(cancel(&record) == -1);
	assert(observed_start == &record);
	assert(observed_finish == &record);
	assert(observed_cancel == &record);
	assert(observed_start->marker == UINTPTR_MAX);

	assert(start(NULL) == INT32_MIN);
	assert(finish(NULL) == INT32_MAX);
	assert(cancel(NULL) == -1);
	assert(observed_start == NULL);
	assert(observed_finish == NULL);
	assert(observed_cancel == NULL);
	return 0;
}
