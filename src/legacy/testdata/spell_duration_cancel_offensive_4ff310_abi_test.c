#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_cancel_offensive_4ff310.h"

typedef void (*spell_duration_cancel_offensive_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF310 caster must remain native-width");
_Static_assert(
	_Generic(&sub_4FF310, spell_duration_cancel_offensive_fn: 1, default: 0),
	"004FF310 must preserve its exact native-pointer ABI");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_caster;
static unsigned int observed_calls;

void sub_4FF310(nox_object_t* caster) {
	observed_caster = caster;
	++observed_calls;
}

int main(void) {
	nox_object_t caster = {.marker = UINTPTR_MAX};
	spell_duration_cancel_offensive_fn const cancel = sub_4FF310;

	cancel(&caster);
	assert(observed_caster == &caster);
	assert(caster.marker == UINTPTR_MAX);
	cancel(NULL);
	assert(observed_caster == NULL);
	assert(observed_calls == 2);
	return 0;
}
