#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_order_533900.h"

typedef void (*unit_order_fn)(nox_object_t*, nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"unit orders must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_orderUnit_533900, unit_order_fn: 1, default: 0),
	"00533900 must receive two native object pointers and one signed dword");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_owner;
static nox_object_t* observed_creature;
static int32_t observed_order;

void nox_xxx_orderUnit_533900(
	nox_object_t* owner, nox_object_t* creature, int32_t order_type
) {
	observed_owner = owner;
	observed_creature = creature;
	observed_order = order_type;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	nox_object_t creature = {.marker = UINTPTR_MAX - 1};
	unit_order_fn const order = nox_xxx_orderUnit_533900;

	order(&owner, &creature, INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_creature == &creature);
	assert(observed_order == INT32_MIN);
	assert(observed_owner->marker == UINTPTR_MAX);
	assert(observed_creature->marker == UINTPTR_MAX - 1);

	order(NULL, NULL, INT32_MAX);
	assert(observed_owner == NULL);
	assert(observed_creature == NULL);
	assert(observed_order == INT32_MAX);
	return 0;
}
