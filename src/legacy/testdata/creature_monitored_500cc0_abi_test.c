#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../creature_monitored_500cc0.h"

typedef int32_t (*creature_monitored_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"00500CC0 results must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	sizeof(nox_xxx_creatureIsMonitored_500CC0(NULL, NULL)) == 4,
	"00500CC0 call expressions must return exactly four bytes");
_Static_assert(
	_Generic(&nox_xxx_creatureIsMonitored_500CC0,
		creature_monitored_fn: 1,
		default: 0),
	"00500CC0 must receive two native object pointers and return int32_t");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_owner;
static nox_object_t* observed_unit;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_creatureIsMonitored_500CC0(
	nox_object_t* owner,
	nox_object_t* unit) {
	observed_owner = owner;
	observed_unit = unit;
	++observed_calls;
	return next_result;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	nox_object_t unit = {.marker = UINTPTR_MAX - 1};
	creature_monitored_fn const monitored =
		nox_xxx_creatureIsMonitored_500CC0;

	next_result = INT32_MIN;
	assert(monitored(&owner, &unit) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_unit == &unit);
	assert(observed_owner->marker == UINTPTR_MAX);
	assert(observed_unit->marker == UINTPTR_MAX - 1);

	next_result = INT32_MAX;
	assert(monitored(NULL, NULL) == INT32_MAX);
	assert(observed_owner == NULL);
	assert(observed_unit == NULL);
	assert(observed_calls == 2);
	return 0;
}
