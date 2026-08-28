// Freestanding C11 matrix companion for GAME.EXE 004F3CE0. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, live flag reload, delayed argument loads, full
// signed DefaultPickup result, and fixed sound 826 behavior.
#include "../pickup_abilitybook_4f3ce0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_abilitybook_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_flags;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(uint32_t) == 4 && sizeof(int32_t) == 4,
	"object fields and callback values must remain exact 32-bit integers");
_Static_assert(
	_Generic(&nox_xxx_pickupAbilitybook_4F3CE0,
		pickup_abilitybook_fn: 1, default: 0),
	"AbilityBookPickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[16];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t game_flags_result;
static int32_t default_result;
static int mutate_destroy_on_use;

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
	game_flags_result = INT32_C(0);
	default_result = INT32_C(0);
	mutate_destroy_on_use = 0;
}

static int32_t game_flags_check(uint32_t flags) {
	event(1U);
	assert(flags == UINT32_C(0x1800));
	return game_flags_result;
}

static void use_by_net_code(nox_object_t* owner, nox_object_t* item) {
	event(2U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	if (mutate_destroy_on_use != 0)
		item->object_flags |= UINT32_C(0x20);
}

static uint8_t load_flags_low(nox_object_t* item) {
	event(3U);
	return (uint8_t)item->object_flags;
}

static int32_t load_arg4(int32_t value) {
	event(4U);
	return value;
}

static int32_t load_arg3(int32_t value) {
	event(5U);
	return value;
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(6U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

static void audio(uint32_t sound, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(7U);
	assert(sound == UINT32_C(826));
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
}

int32_t nox_xxx_pickupAbilitybook_4F3CE0(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	if (game_flags_check(UINT32_C(0x1800)) != INT32_C(0))
		use_by_net_code(owner, item);
	if ((load_flags_low(item) & UINT8_C(0x20)) != UINT8_C(0))
		return INT32_C(1);

	int32_t loaded_arg4 = load_arg4(arg4);
	int32_t loaded_arg3 = load_arg3(arg3);
	int32_t result = default_pickup(owner, item, loaded_arg3, loaded_arg4);
	if (result != INT32_C(0))
		audio(UINT32_C(826), owner, INT32_C(0), UINT32_C(0));
	return result;
}

static void assert_trace(const unsigned int* want, size_t count) {
	assert(trace_count == count);
	for (size_t index = 0; index < count; ++index)
		assert(trace[index] == want[index]);
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX - (uintptr_t)1};
	nox_object_t item = {.marker = UINTPTR_MAX - (uintptr_t)2};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}

	reset_case(&owner, &item);
	game_flags_result = INT32_C(-1);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_result = INT32_MIN;
	assert(nox_xxx_pickupAbilitybook_4F3CE0(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	{
		const unsigned int want[] = {1U, 2U, 3U, 4U, 5U, 6U, 7U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	item.object_flags = UINT32_C(0);
	reset_case(&owner, &item);
	game_flags_result = INT32_C(1);
	mutate_destroy_on_use = 1;
	assert(nox_xxx_pickupAbilitybook_4F3CE0(
		&owner, &item, INT32_C(3), INT32_C(4)) == INT32_C(1));
	{
		const unsigned int want[] = {1U, 2U, 3U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	item.object_flags = UINT32_C(0);
	reset_case(&owner, &item);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	assert(nox_xxx_pickupAbilitybook_4F3CE0(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(0));
	{
		const unsigned int want[] = {1U, 3U, 4U, 5U, 6U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	item.object_flags = UINT32_C(0x120);
	reset_case(&owner, &item);
	assert(nox_xxx_pickupAbilitybook_4F3CE0(
		&owner, &item, INT32_C(9), INT32_C(10)) == INT32_C(1));
	{
		const unsigned int want[] = {1U, 3U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	item.object_flags = UINT32_C(0);
	reset_case(&owner, &item);
	expected_arg3 = INT32_C(9);
	expected_arg4 = INT32_C(10);
	default_result = INT32_MAX;
	assert(nox_xxx_pickupAbilitybook_4F3CE0(
		&owner, &item, INT32_C(9), INT32_C(10)) == INT32_MAX);
	{
		const unsigned int want[] = {1U, 3U, 4U, 5U, 6U, 7U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	return 0;
}
