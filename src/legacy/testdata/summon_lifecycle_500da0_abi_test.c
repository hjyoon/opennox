#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../summon_lifecycle_500da0.h"

typedef int32_t (*summon_phase_fn)(nox_dur_spell_t*);
typedef void (*summon_cancel_fn)(nox_dur_spell_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"summon start/finish results must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_summonStart_500DA0, summon_phase_fn: 1, default: 0),
	"00500DA0 must receive one native duration-record pointer");
_Static_assert(
	_Generic(&nox_xxx_summonFinish_5010D0, summon_phase_fn: 1, default: 0),
	"005010D0 must receive one native duration-record pointer");
_Static_assert(
	_Generic(&nox_xxx_summonCancel_5011C0, summon_cancel_fn: 1, default: 0),
	"005011C0 must receive one native duration-record pointer");

struct nox_dur_spell_t {
	uintptr_t marker;
};

static nox_dur_spell_t* observed_start;
static nox_dur_spell_t* observed_finish;
static nox_dur_spell_t* observed_cancel;

int32_t nox_xxx_summonStart_500DA0(nox_dur_spell_t* record) {
	observed_start = record;
	return INT32_MIN;
}

int32_t nox_xxx_summonFinish_5010D0(nox_dur_spell_t* record) {
	observed_finish = record;
	return INT32_MAX;
}

void nox_xxx_summonCancel_5011C0(nox_dur_spell_t* record) {
	observed_cancel = record;
}

int main(void) {
	nox_dur_spell_t record = {.marker = UINTPTR_MAX};
	summon_phase_fn const start = nox_xxx_summonStart_500DA0;
	summon_phase_fn const finish = nox_xxx_summonFinish_5010D0;
	summon_cancel_fn const cancel = nox_xxx_summonCancel_5011C0;

	assert(start(&record) == INT32_MIN);
	assert(finish(&record) == INT32_MAX);
	cancel(&record);
	assert(observed_start == &record);
	assert(observed_finish == &record);
	assert(observed_cancel == &record);
	assert(observed_start->marker == UINTPTR_MAX);

	assert(start(NULL) == INT32_MIN);
	assert(finish(NULL) == INT32_MAX);
	cancel(NULL);
	assert(observed_start == NULL);
	assert(observed_finish == NULL);
	assert(observed_cancel == NULL);
	return 0;
}
