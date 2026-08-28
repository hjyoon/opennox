// Freestanding C11 matrix companion for GAME.EXE 004F37D0. Production uses
// the Go restoration; this independently locks the four-argument ABI, native
// pointer width, eager PotionUseData value, class-scaled health/mana use,
// cure deletion, decay/default fallback, exact result, and pickup-audio gate.
#include "../pickup_potion_4f37d0.h"

#include <assert.h>
#include <limits.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_potion_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

typedef struct potion_use_data {
	int32_t value;
} potion_use_data;

typedef struct health_data {
	uint16_t current;
	uint16_t unused;
	uint16_t maximum;
} health_data;

typedef struct player_data {
	uint8_t class_value;
} player_data;

typedef struct player_update_data {
	uint16_t unused0;
	uint16_t unused2;
	uint16_t mana_current;
	uint16_t mana_previous;
	uint16_t mana_maximum;
	player_data* player;
} player_update_data;

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	uint32_t subclass;
	uint32_t net_code;
	uint8_t poison;
	health_data* health;
	player_update_data* update;
	potion_use_data* use_data;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4 && sizeof(float) == 4,
	"callback scalars and class multipliers must remain 32-bit values");
_Static_assert(sizeof(potion_use_data) == 4 &&
	offsetof(potion_use_data, value) == 0,
	"PotionUseData must retain its leading signed dword");
_Static_assert(offsetof(health_data, current) == 0 &&
	offsetof(health_data, maximum) == 4,
	"HealthData current and maximum words must retain their ABI32 positions");
_Static_assert(offsetof(player_update_data, mana_current) == 4 &&
	offsetof(player_update_data, mana_maximum) == 8,
	"PlayerUpdateData mana words must retain their ABI32 positions");
_Static_assert(
	_Generic(&nox_xxx_pickupPotion_4F37D0,
		pickup_potion_fn: 1, default: 0),
	"PotionPickup must retain two native pointers, two int32_t values, and int32_t result");

#define CLASS_PLAYER UINT32_C(0x04)
#define SUBCLASS_HEALTH UINT32_C(0x10)
#define SUBCLASS_MANA UINT32_C(0x20)
#define SUBCLASS_CURE UINT32_C(0x40)

static unsigned int trace[64];
static size_t trace_count;
static int32_t class_restriction_flag;
static int32_t quest_flag;
static int32_t player_state_value;
static int32_t class_can_use_value;
static float health_multiplier[3];
static float mana_multiplier[3];
static int32_t default_value;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t adjusted_health;
static int32_t added_mana;
static uint32_t played_sound[4];
static size_t played_sound_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_potion;
static int mutate_class_after_health;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void expect_trace(const unsigned int* values, size_t count) {
	assert(trace_count == count);
	for (size_t i = 0U; i < count; ++i)
		assert(trace[i] == values[i]);
}

static void reset_case(nox_object_t* owner, nox_object_t* potion) {
	trace_count = 0U;
	class_restriction_flag = INT32_C(0);
	quest_flag = INT32_C(0);
	player_state_value = INT32_C(0);
	class_can_use_value = INT32_C(1);
	health_multiplier[0] = 1.0F;
	health_multiplier[1] = 1.0F;
	health_multiplier[2] = 1.0F;
	mana_multiplier[0] = 1.0F;
	mana_multiplier[1] = 1.0F;
	mana_multiplier[2] = 1.0F;
	default_value = INT32_C(0);
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	adjusted_health = INT32_C(0);
	added_mana = INT32_C(0);
	played_sound_count = 0U;
	expected_owner = owner;
	expected_potion = potion;
	mutate_class_after_health = 0;
}

static int32_t game_flag(uint32_t flag) {
	event(flag == UINT32_C(0x2000) ? 1U : 2U);
	return flag == UINT32_C(0x2000)
		? class_restriction_flag
		: quest_flag;
}

static int32_t player_class_can_use(nox_object_t* potion, uint8_t class_value) {
	event(3U);
	assert(potion == expected_potion);
	assert(class_value == expected_owner->update->player->class_value);
	return class_can_use_value;
}

static void class_failure(nox_object_t* owner) {
	event(4U);
	assert(owner == expected_owner);
	owner->net_code = UINT32_C(0x89ABCDEF);
}

static int32_t player_state(nox_object_t* owner) {
	event(5U);
	assert(owner == expected_owner);
	return player_state_value;
}

static void audio(
		uint32_t sound,
		nox_object_t* owner,
		int32_t kind,
		uint32_t code) {
	event(6U);
	assert(owner == expected_owner);
	if (sound == UINT32_C(925)) {
		assert(kind == INT32_C(2));
		assert(code == UINT32_C(0x89ABCDEF));
	} else {
		assert(kind == INT32_C(0));
		assert(code == UINT32_C(0));
	}
	assert(played_sound_count < sizeof(played_sound) / sizeof(played_sound[0]));
	played_sound[played_sound_count++] = sound;
}

static int32_t float_to_int(float value) {
	double const rounded = nearbyint((double)value);
	if (!isfinite(rounded) ||
		rounded < (double)INT32_MIN || rounded > (double)INT32_MAX)
		return INT32_MIN;
	return (int32_t)rounded;
}

static int32_t scale_amount(int32_t base, float multiplier) {
	float const product = (float)((double)base * (double)multiplier);
	return float_to_int(product);
}

static void adjust_health(nox_object_t* owner, int32_t amount) {
	event(7U);
	assert(owner == expected_owner);
	adjusted_health = amount;
	if (mutate_class_after_health)
		owner->update->player->class_value = UINT8_C(7);
}

static void add_mana(nox_object_t* owner, int32_t amount) {
	event(8U);
	assert(owner == expected_owner);
	added_mana = amount;
}

static void remove_poison(nox_object_t* owner) {
	event(9U);
	assert(owner == expected_owner);
	owner->poison = UINT8_C(0);
}

static uint32_t spell_audio(int32_t spell, int32_t field) {
	event(10U);
	assert(spell == INT32_C(14));
	assert(field == INT32_C(1));
	return UINT32_C(0x456);
}

static void delayed_delete(nox_object_t* potion) {
	event(11U);
	assert(potion == expected_potion);
}

static void decay(nox_object_t* potion) {
	event(12U);
	assert(potion == expected_potion);
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* potion,
		int32_t arg3,
		int32_t arg4) {
	event(13U);
	assert(owner == expected_owner);
	assert(potion == expected_potion);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_value;
}

int32_t nox_xxx_pickupPotion_4F37D0(
		nox_object_t* owner,
		nox_object_t* potion,
		int32_t arg3,
		int32_t arg4) {
	int32_t const base_amount = potion->use_data->value;
	int32_t amount = base_amount;
	int consumed = 0;

	if (game_flag(UINT32_C(0x2000)) != INT32_C(0) &&
		game_flag(UINT32_C(0x1000)) == INT32_C(0) &&
		((uint8_t)owner->object_class & UINT8_C(0x04)) != UINT8_C(0)) {
		uint8_t const class_value = owner->update->player->class_value;
		if (player_class_can_use(potion, class_value) == INT32_C(0)) {
			class_failure(owner);
			audio(UINT32_C(925), owner, INT32_C(2), owner->net_code);
			return INT32_C(0);
		}
	}

	if (player_state(owner) == INT32_C(0)) {
		if (((uint8_t)potion->subclass & UINT8_C(0x10)) != UINT8_C(0)) {
			health_data* health = owner->health;
			if (health != NULL) {
				if (((uint8_t)owner->object_class & UINT8_C(0x04)) != UINT8_C(0)) {
					uint8_t const class_value = owner->update->player->class_value;
					if (class_value <= UINT8_C(2))
						amount = scale_amount(base_amount, health_multiplier[class_value]);
				}
				health = owner->health;
				if ((int32_t)((uint32_t)health->current + (uint32_t)amount) <
					(int32_t)health->maximum) {
					adjust_health(owner, amount);
					audio(UINT32_C(754), owner, INT32_C(0), UINT32_C(0));
					consumed = 1;
				}
			}
		}

		if (((uint8_t)potion->subclass & UINT8_C(0x20)) != UINT8_C(0) &&
			((uint8_t)owner->object_class & UINT8_C(0x04)) != UINT8_C(0)) {
			player_update_data* const update = owner->update;
			uint8_t const class_value = update->player->class_value;
			if (class_value <= UINT8_C(2))
				amount = scale_amount(base_amount, mana_multiplier[class_value]);
			if ((int32_t)((uint32_t)update->mana_current + (uint32_t)amount) <
				(int32_t)update->mana_maximum) {
				add_mana(owner, amount);
				audio(UINT32_C(755), owner, INT32_C(0), UINT32_C(0));
				consumed = 1;
			}
		}

		if (((uint8_t)potion->subclass & UINT8_C(0x40)) != UINT8_C(0) &&
			((uint8_t)owner->object_class & UINT8_C(0x04)) != UINT8_C(0) &&
			owner->poison != UINT8_C(0)) {
			remove_poison(owner);
			audio(spell_audio(INT32_C(14), INT32_C(1)), owner, INT32_C(0), UINT32_C(0));
			delayed_delete(potion);
			return INT32_C(1);
		}

		if (consumed != 0) {
			delayed_delete(potion);
			return INT32_C(1);
		}
	}

	decay(potion);
	int32_t const result = default_pickup(owner, potion, arg3, arg4);
	if (result == INT32_C(1))
		audio(UINT32_C(832), owner, INT32_C(0), UINT32_C(0));
	return result;
}

static void test_class_rejection(void) {
	player_data player = {.class_value = UINT8_C(1)};
	player_update_data update = {.player = &player};
	potion_use_data use = {.value = INT32_C(5)};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = CLASS_PLAYER,
		.net_code = UINT32_C(7),
		.update = &update,
	};
	nox_object_t potion = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.use_data = &use,
	};
	static const unsigned int want[] = {1U, 2U, 3U, 4U, 6U};

	reset_case(&owner, &potion);
	class_restriction_flag = INT32_MIN;
	class_can_use_value = INT32_C(0);
	assert(nox_xxx_pickupPotion_4F37D0(
		&owner, &potion, INT32_MIN, INT32_MAX) == INT32_C(0));
	expect_trace(want, sizeof(want) / sizeof(want[0]));
	assert(played_sound_count == 1U && played_sound[0] == UINT32_C(925));
}

static void test_health_mana_cure(void) {
	player_data player = {.class_value = UINT8_C(1)};
	player_update_data update = {
		.mana_current = UINT16_C(3),
		.mana_maximum = UINT16_C(100),
		.player = &player,
	};
	health_data health = {
		.current = UINT16_C(4),
		.maximum = UINT16_C(100),
	};
	potion_use_data use = {.value = INT32_C(5)};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = CLASS_PLAYER,
		.poison = UINT8_C(1),
		.health = &health,
		.update = &update,
	};
	nox_object_t potion = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.subclass = SUBCLASS_HEALTH | SUBCLASS_MANA | SUBCLASS_CURE,
		.use_data = &use,
	};
	static const unsigned int want[] = {
		1U, 5U, 7U, 6U, 8U, 6U, 9U, 10U, 6U, 11U,
	};

	reset_case(&owner, &potion);
	health_multiplier[1] = 0.5F;
	mana_multiplier[1] = 1.5F;
	assert(nox_xxx_pickupPotion_4F37D0(
		&owner, &potion, INT32_MIN, INT32_MAX) == INT32_C(1));
	expect_trace(want, sizeof(want) / sizeof(want[0]));
	assert(adjusted_health == INT32_C(2));
	assert(added_mana == INT32_C(8));
	assert(owner.poison == UINT8_C(0));
	assert(played_sound_count == 3U);
	assert(played_sound[0] == UINT32_C(754));
	assert(played_sound[1] == UINT32_C(755));
	assert(played_sound[2] == UINT32_C(0x456));
}

static void test_invalid_mana_class_retains_health_amount(void) {
	player_data player = {.class_value = UINT8_C(0)};
	player_update_data update = {
		.mana_current = UINT16_C(1),
		.mana_maximum = UINT16_C(100),
		.player = &player,
	};
	health_data health = {
		.current = UINT16_C(1),
		.maximum = UINT16_C(100),
	};
	potion_use_data use = {.value = INT32_C(10)};
	nox_object_t owner = {
		.object_class = CLASS_PLAYER,
		.health = &health,
		.update = &update,
	};
	nox_object_t potion = {
		.subclass = SUBCLASS_HEALTH | SUBCLASS_MANA,
		.use_data = &use,
	};

	reset_case(&owner, &potion);
	health_multiplier[0] = 3.0F;
	mutate_class_after_health = 1;
	assert(nox_xxx_pickupPotion_4F37D0(
		&owner, &potion, INT32_C(1), INT32_C(2)) == INT32_C(1));
	assert(adjusted_health == INT32_C(30));
	assert(added_mana == INT32_C(30));
}

static void test_default_exact_result_and_audio_gate(void) {
	player_data player = {.class_value = UINT8_C(0)};
	player_update_data update = {.player = &player};
	potion_use_data use = {.value = INT32_C(-10)};
	nox_object_t owner = {.update = &update};
	nox_object_t potion = {.use_data = &use};
	static const unsigned int want_no_audio[] = {1U, 5U, 12U, 13U};
	static const unsigned int want_audio[] = {1U, 5U, 12U, 13U, 6U};

	reset_case(&owner, &potion);
	player_state_value = INT32_MIN;
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_value = INT32_MIN;
	assert(nox_xxx_pickupPotion_4F37D0(
		&owner, &potion, INT32_MIN, INT32_MAX) == INT32_MIN);
	expect_trace(want_no_audio, sizeof(want_no_audio) / sizeof(want_no_audio[0]));
	assert(played_sound_count == 0U);

	reset_case(&owner, &potion);
	player_state_value = INT32_C(1);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	default_value = INT32_C(1);
	assert(nox_xxx_pickupPotion_4F37D0(
		&owner, &potion, INT32_C(-17), INT32_C(-23)) == INT32_C(1));
	expect_trace(want_audio, sizeof(want_audio) / sizeof(want_audio[0]));
	assert(played_sound_count == 1U && played_sound[0] == UINT32_C(832));
}

static void test_rounding_and_signed_wrap(void) {
	assert(scale_amount(INT32_C(3), 0.5F) == INT32_C(2));
	assert(scale_amount(INT32_C(5), 0.5F) == INT32_C(2));
	assert(scale_amount(INT32_C(-3), 0.5F) == INT32_C(-2));
	assert(scale_amount(INT32_C(-5), 0.5F) == INT32_C(-2));
	assert(scale_amount(INT32_C(16777217), 1.0F) == INT32_C(16777216));
	assert(scale_amount(INT32_MAX, 1.0F) == INT32_MIN);
	assert(scale_amount(INT32_C(1), INFINITY) == INT32_MIN);
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t potion = {0};
	if (sizeof(void*) == 8U) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&potion > UINT32_MAX);
	}

	test_class_rejection();
	test_health_mana_cure();
	test_invalid_mana_class_retains_health_amount();
	test_default_exact_result_and_audio_gate();
	test_rounding_and_signed_wrap();
	return 0;
}
