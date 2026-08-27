// Freestanding C11 matrix companion for GAME.EXE 004F34D0. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, unconditional use helper, live Destroyed gate,
// and DefaultPickup forwarding/result behavior.
#include "../pickup_use_4f34d0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_use_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef int32_t (*use_fn)(nox_object_t*, nox_object_t*);

struct nox_object_t {
	uintptr_t marker;
	uint32_t flags;
	use_fn use;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"callback scalars and result must remain exact 32-bit integers");
_Static_assert(
	_Generic(&nox_xxx_pickupUse_4F34D0,
		pickup_use_fn: 1, default: 0),
	"UsePickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[16];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t use_result;
static int32_t default_result;
static int32_t special_state;
static int destroy_in_use;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	trace_count = 0U;
	expected_owner = owner;
	expected_item = item;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	use_result = INT32_C(0);
	default_result = INT32_C(0);
	special_state = INT32_C(0);
	destroy_in_use = 0;
}

static int32_t use_item(nox_object_t* owner, nox_object_t* item) {
	event(1U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	if (destroy_in_use)
		item->flags |= UINT32_C(0x20);
	return use_result;
}

static int32_t use_by_net_code(nox_object_t* owner, nox_object_t* item) {
	event(2U);
	if (item == NULL || item->use == NULL)
		return INT32_C(1);
	if (special_state == INT32_C(1))
		return special_state;
	return item->use(owner, item);
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(3U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

int32_t nox_xxx_pickupUse_4F34D0(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	(void)use_by_net_code(owner, item);
	if (((uint8_t)item->flags & UINT8_C(0x20)) != UINT8_C(0))
		return INT32_C(1);
	return default_pickup(owner, item, arg3, arg4);
}

int main(void) {
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.use = use_item,
	};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}

	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	use_result = INT32_MIN;
	default_result = INT32_MAX;
	assert(nox_xxx_pickupUse_4F34D0(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MAX);
	assert(trace_count == 3U);
	assert(trace[0] == 2U && trace[1] == 1U && trace[2] == 3U);

	item.flags = UINT32_C(0x80000000);
	reset_case(&owner, &item);
	destroy_in_use = 1;
	use_result = INT32_MIN;
	default_result = INT32_MIN;
	assert(nox_xxx_pickupUse_4F34D0(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(1));
	assert(trace_count == 2U && trace[0] == 2U && trace[1] == 1U);
	assert((item.flags & UINT32_C(0x20)) != UINT32_C(0));

	item.flags = UINT32_C(0);
	item.use = NULL;
	reset_case(&owner, &item);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	default_result = INT32_MIN;
	assert(nox_xxx_pickupUse_4F34D0(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_MIN);
	assert(trace_count == 2U && trace[0] == 2U && trace[1] == 3U);

	item.use = use_item;
	reset_case(&owner, &item);
	special_state = INT32_C(1);
	default_result = INT32_C(-1);
	assert(nox_xxx_pickupUse_4F34D0(
		&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(-1));
	assert(trace_count == 2U && trace[0] == 2U && trace[1] == 3U);

	return 0;
}
