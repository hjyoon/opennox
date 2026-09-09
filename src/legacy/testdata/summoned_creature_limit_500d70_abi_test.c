#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../summoned_creature_limit_500d70.h"

typedef int32_t (*summoned_creature_limit_fn)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"00500D70 guide index and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	sizeof(nox_xxx_checkSummonedCreaturesLimit_500D70(NULL, 0)) == 4,
	"00500D70 call expressions must return exactly four bytes");
_Static_assert(
	_Generic(&nox_xxx_checkSummonedCreaturesLimit_500D70,
		summoned_creature_limit_fn: 1,
		default: 0),
	"00500D70 must receive a native object pointer and signed dword index");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_owner;
static int32_t observed_guide_index;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_checkSummonedCreaturesLimit_500D70(
	nox_object_t* owner,
	int32_t guide_index) {
	observed_owner = owner;
	observed_guide_index = guide_index;
	++observed_calls;
	return next_result;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	summoned_creature_limit_fn const check_limit =
		nox_xxx_checkSummonedCreaturesLimit_500D70;

	next_result = 1;
	assert(check_limit(&owner, INT32_MIN) == 1);
	assert(observed_owner == &owner);
	assert(observed_owner->marker == UINTPTR_MAX);
	assert(observed_guide_index == INT32_MIN);

	next_result = 0;
	assert(check_limit(NULL, INT32_MAX) == 0);
	assert(observed_owner == NULL);
	assert(observed_guide_index == INT32_MAX);
	assert(observed_calls == 2);
	return 0;
}
