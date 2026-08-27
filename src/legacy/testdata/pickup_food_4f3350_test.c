// Freestanding C11 matrix companion for GAME.EXE 004F3350. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, Use/Destroyed gates, DefaultPickup forwarding,
// exact sound-table layout, and subclass-before-material scan.
#include "../pickup_food_4f3350.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_food_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef int32_t (*use_fn)(nox_object_t*, nox_object_t*);

typedef struct sound_rule {
	uint32_t subclass_mask;
	uint16_t material_mask;
	uint16_t sound;
} sound_rule;

struct nox_object_t {
	uintptr_t marker;
	uint32_t subclass;
	uint32_t flags;
	uint16_t material;
	use_fn use;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"callback scalars and result must remain exact 32-bit integers");
_Static_assert(sizeof(sound_rule) == 8,
	"FoodPickup sound rows must remain exact eight-byte records");
_Static_assert(offsetof(sound_rule, subclass_mask) == 0,
	"subclass mask must remain the first dword");
_Static_assert(offsetof(sound_rule, material_mask) == 4,
	"material mask must remain the following word");
_Static_assert(offsetof(sound_rule, sound) == 6,
	"sound must remain the final word");
_Static_assert(
	_Generic(&nox_xxx_pickupFood_4F3350,
		pickup_food_fn: 1, default: 0),
	"FoodPickup must retain two native pointers, two int32_t values, and int32_t result");

static sound_rule rules[5] = {
	{UINT32_C(0x00000000), UINT16_C(0x0001), UINT16_C(834)},
	{UINT32_C(0x00000002), UINT16_C(0x0000), UINT16_C(836)},
	{UINT32_C(0x00000004), UINT16_C(0x0000), UINT16_C(832)},
	{UINT32_C(0x00000080), UINT16_C(0x0000), UINT16_C(838)},
	{UINT32_C(0x00000000), UINT16_C(0x0000), UINT16_C(0)},
};
static const sound_rule oracle_rules[5] = {
	{UINT32_C(0x00000000), UINT16_C(0x0001), UINT16_C(834)},
	{UINT32_C(0x00000002), UINT16_C(0x0000), UINT16_C(836)},
	{UINT32_C(0x00000004), UINT16_C(0x0000), UINT16_C(832)},
	{UINT32_C(0x00000080), UINT16_C(0x0000), UINT16_C(838)},
	{UINT32_C(0x00000000), UINT16_C(0x0000), UINT16_C(0)},
};

static unsigned int trace[64];
static size_t trace_count;
static int32_t player_state_value;
static int32_t default_value;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static uint32_t played_sound;
static nox_object_t* played_owner;
static int32_t use_result;
static int destroy_in_use;
static int mutate_after_default;
static int mutate_matched_sound;
static unsigned int row_one_sound_loads;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	for (size_t i = 0; i < 5; ++i)
		rules[i] = oracle_rules[i];
	trace_count = 0;
	player_state_value = INT32_C(0);
	default_value = INT32_C(1);
	expected_owner = owner;
	expected_item = item;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	played_sound = UINT32_C(0);
	played_owner = NULL;
	use_result = INT32_C(0);
	destroy_in_use = 0;
	mutate_after_default = 0;
	mutate_matched_sound = 0;
	row_one_sound_loads = 0U;
}

static int32_t player_state(nox_object_t* owner) {
	event(1U);
	assert(owner == expected_owner);
	return player_state_value;
}

static int32_t use_item(nox_object_t* owner, nox_object_t* item) {
	event(2U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	if (destroy_in_use)
		item->flags |= UINT32_C(0x20);
	return use_result;
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
	if (mutate_after_default)
		item->subclass = UINT32_C(0x00000002);
	return default_value;
}

static uint16_t load_sound(size_t row) {
	event((unsigned int)(10U + row));
	uint16_t const value = rules[row].sound;
	if (row == 1U) {
		++row_one_sound_loads;
		if (mutate_matched_sound && row_one_sound_loads == 1U)
			rules[1].sound = UINT16_C(901);
	}
	return value;
}

static void audio(uint32_t sound, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(20U);
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
	played_sound = sound;
	played_owner = owner;
}

int32_t nox_xxx_pickupFood_4F3350(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	if (owner == NULL || item == NULL)
		return INT32_C(0);

	if (player_state(owner) == INT32_C(0) &&
		((uint8_t)item->subclass & UINT8_C(0x84)) == UINT8_C(0)) {
		use_fn const use = item->use;
		(void)use(owner, item);
	}

	if (((uint8_t)item->flags & UINT8_C(0x20)) != UINT8_C(0))
		return INT32_C(1);

	int32_t const result = default_pickup(owner, item, arg3, arg4);
	if (result == INT32_C(0))
		return result;

	size_t row = 0U;
	if (load_sound(row) == UINT16_C(0))
		return result;
	for (;;) {
		uint32_t const subclass = item->subclass;
		if ((subclass & rules[row].subclass_mask) != UINT32_C(0)) {
			uint16_t const sound = load_sound(row);
			audio((uint32_t)sound, owner, INT32_C(0), UINT32_C(0));
			return result;
		}

		uint16_t const material_mask = rules[row].material_mask;
		if ((material_mask & item->material) != UINT16_C(0)) {
			uint16_t const sound = load_sound(row);
			audio((uint32_t)sound, owner, INT32_C(0), UINT32_C(0));
			return result;
		}

		++row;
		if (load_sound(row) == UINT16_C(0))
			return result;
	}
}

int main(void) {
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.use = use_item,
	};

	reset_case(&owner, &item);
	assert(nox_xxx_pickupFood_4F3350(NULL, &item, INT32_C(1), INT32_C(2)) == INT32_C(0));
	assert(nox_xxx_pickupFood_4F3350(&owner, NULL, INT32_C(1), INT32_C(2)) == INT32_C(0));
	assert(trace_count == 0U);

	reset_case(&owner, &item);
	destroy_in_use = 1;
	use_result = INT32_MIN;
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(7), INT32_C(9)) == INT32_C(1));
	assert(trace_count == 2U && trace[0] == 1U && trace[1] == 2U);
	assert((item.flags & UINT32_C(0x20)) != UINT32_C(0));

	item.flags = UINT32_C(0);
	item.subclass = UINT32_C(0x00000004);
	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_value = INT32_MIN;
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(played_sound == UINT32_C(832) && played_owner == &owner);
	assert(trace_count > 2U && trace[0] == 1U && trace[1] == 3U);

	item.subclass = UINT32_C(0);
	reset_case(&owner, &item);
	player_state_value = INT32_MIN;
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	default_value = INT32_MAX;
	mutate_after_default = 1;
	mutate_matched_sound = 1;
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_MAX);
	assert(row_one_sound_loads == 2U);
	assert(played_sound == UINT32_C(901));

	item.subclass = UINT32_C(0x00000004);
	item.material = UINT16_C(0x0001);
	reset_case(&owner, &item);
	player_state_value = INT32_C(1);
	default_value = INT32_C(-1);
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(-1));
	assert(played_sound == UINT32_C(834));

	item.subclass = UINT32_C(0);
	item.material = UINT16_C(0);
	reset_case(&owner, &item);
	player_state_value = INT32_C(1);
	default_value = INT32_MAX;
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(0), INT32_C(0)) == INT32_MAX);
	assert(played_owner == NULL);
	assert(trace[trace_count - 1U] == 14U);

	reset_case(&owner, &item);
	player_state_value = INT32_C(1);
	default_value = INT32_C(-1);
	rules[0].sound = UINT16_C(0);
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(-1));
	assert(trace_count == 3U && trace[2] == 10U);

	reset_case(&owner, &item);
	player_state_value = INT32_C(1);
	default_value = INT32_C(0);
	assert(nox_xxx_pickupFood_4F3350(&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(0));
	assert(trace_count == 2U && trace[1] == 3U);

	return 0;
}
