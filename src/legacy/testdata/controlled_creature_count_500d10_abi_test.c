#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../controlled_creature_count_500d10.h"

typedef int32_t (*controlled_creature_count_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"00500D10 results must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	sizeof(nox_xxx_countControlledCreatures_500D10(NULL)) == 4,
	"00500D10 call expressions must return exactly four bytes");
_Static_assert(
	_Generic(&nox_xxx_countControlledCreatures_500D10,
		controlled_creature_count_fn: 1,
		default: 0),
	"00500D10 must receive one native object pointer and return int32_t");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_owner;
static int32_t next_result;
static unsigned int observed_calls;

int32_t nox_xxx_countControlledCreatures_500D10(nox_object_t* owner) {
	observed_owner = owner;
	++observed_calls;
	return next_result;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	controlled_creature_count_fn const count =
		nox_xxx_countControlledCreatures_500D10;

	next_result = INT32_MIN;
	assert(count(&owner) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_owner->marker == UINTPTR_MAX);

	next_result = INT32_MAX;
	assert(count(NULL) == INT32_MAX);
	assert(observed_owner == NULL);
	assert(observed_calls == 2);
	return 0;
}
