// Freestanding C11 matrix companion for GAME.EXE 004F3510. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, owner-chain gate, full int32 result, and ordered
// success/rejection audio behavior.
#include "../pickup_trap_4f3510.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_trap_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	uint32_t net_code;
	nox_object_t* owner;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"callback scalars and result must remain exact 32-bit integers");
_Static_assert(
	_Generic(&nox_xxx_pickupTrap_4F3510,
		pickup_trap_fn: 1, default: 0),
	"TrapPickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[16];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t default_result;

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
	default_result = INT32_C(0);
}

static int32_t has_owner(nox_object_t* item, nox_object_t* owner) {
	event(1U);
	if (item == NULL || owner == NULL)
		return INT32_C(0);
	for (; item != NULL; item = item->owner) {
		if (item == owner)
			return INT32_C(1);
	}
	return INT32_C(0);
}

static int32_t load_arg4(int32_t value) {
	event(2U);
	return value;
}

static int32_t load_arg3(int32_t value) {
	event(3U);
	return value;
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(4U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

static uint8_t load_class_low(nox_object_t* owner) {
	event(6U);
	return (uint8_t)owner->object_class;
}

static uint32_t load_net_code(nox_object_t* owner) {
	event(7U);
	return owner->net_code;
}

static void audio(uint32_t id, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(5U);
	assert(owner == expected_owner);
	if (id == UINT32_C(824)) {
		assert(kind == INT32_C(0));
		assert(code == UINT32_C(0));
	} else {
		assert(id == UINT32_C(925));
		assert(kind == INT32_C(2));
		assert(code == owner->net_code);
	}
}

int32_t nox_xxx_pickupTrap_4F3510(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	if (has_owner(item, owner) != INT32_C(0)) {
		int32_t loaded_arg4 = load_arg4(arg4);
		int32_t loaded_arg3 = load_arg3(arg3);
		int32_t result = default_pickup(
			owner, item, loaded_arg3, loaded_arg4);
		if (result != INT32_C(0))
			audio(UINT32_C(824), owner, INT32_C(0), UINT32_C(0));
		return result;
	}
	if ((load_class_low(owner) & UINT8_C(0x04)) != UINT8_C(0)) {
		uint32_t code = load_net_code(owner);
		audio(UINT32_C(925), owner, INT32_C(2), code);
	}
	return INT32_C(0);
}

int main(void) {
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = UINT32_C(0x80000004),
		.net_code = UINT32_C(0xfedcba98),
	};
	nox_object_t middle = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.owner = &owner,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
		.owner = &middle,
	};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}

	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_result = INT32_MIN;
	assert(nox_xxx_pickupTrap_4F3510(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(trace_count == 5U);
	assert(trace[0] == 1U && trace[1] == 2U && trace[2] == 3U);
	assert(trace[3] == 4U && trace[4] == 5U);

	reset_case(&owner, &item);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	assert(nox_xxx_pickupTrap_4F3510(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(0));
	assert(trace_count == 4U);
	assert(trace[0] == 1U && trace[1] == 2U && trace[2] == 3U && trace[3] == 4U);

	item.owner = NULL;
	reset_case(&owner, &item);
	assert(nox_xxx_pickupTrap_4F3510(
		&owner, &item, INT32_C(7), INT32_C(9)) == INT32_C(0));
	assert(trace_count == 4U);
	assert(trace[0] == 1U && trace[1] == 6U && trace[2] == 7U && trace[3] == 5U);

	owner.object_class = UINT32_C(0x80000000);
	reset_case(&owner, &item);
	assert(nox_xxx_pickupTrap_4F3510(
		&owner, &item, INT32_C(7), INT32_C(9)) == INT32_C(0));
	assert(trace_count == 2U && trace[0] == 1U && trace[1] == 6U);

	owner.object_class = UINT32_C(0x04);
	reset_case(&owner, NULL);
	assert(nox_xxx_pickupTrap_4F3510(
		&owner, NULL, INT32_C(0), INT32_C(0)) == INT32_C(0));
	assert(trace_count == 4U);
	assert(trace[0] == 1U && trace[1] == 6U && trace[2] == 7U && trace[3] == 5U);

	return 0;
}
