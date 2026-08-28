#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_inventory_4f3e30.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*inventory_xfer_fn)(uint16_t, nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(uint16_t) == 2, "map version must remain uint16");
_Static_assert(sizeof(int32_t) == 4, "inventory count and result must remain int32");
_Static_assert(
	_Generic(&nox_xxx_xfer_4F3E30, inventory_xfer_fn: 1, default: 0),
	"inventory transfer must retain uint16, native pointer, int32, and int32 result");

static uint16_t observed_version;
static nox_object_t* observed_owner;
static int32_t observed_count;

int32_t nox_xxx_xfer_4F3E30(uint16_t version, nox_object_t* owner, int32_t count) {
	observed_version = version;
	observed_owner = owner;
	observed_count = count;
	return count;
}

int main(void) {
	nox_object_t owner = {UINTPTR_MAX};
	inventory_xfer_fn const xfer = nox_xxx_xfer_4F3E30;

	assert(xfer(UINT16_MAX, &owner, INT32_MIN) == INT32_MIN);
	assert(observed_version == UINT16_MAX);
	assert(observed_owner == &owner);
	assert(observed_owner->marker == UINTPTR_MAX);
	assert(observed_count == INT32_MIN);

	assert(xfer(UINT16_C(60), NULL, INT32_MAX) == INT32_MAX);
	assert(observed_version == UINT16_C(60));
	assert(observed_owner == NULL);
	assert(observed_count == INT32_MAX);
	return 0;
}
