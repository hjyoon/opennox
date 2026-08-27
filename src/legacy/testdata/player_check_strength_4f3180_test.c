// Freestanding C11 matrix companion for GAME.EXE 004F3180. Production uses
// the Go restoration; this file independently locks the exact public ABI,
// native pointer width, original load/call order, live item state, definition
// selection, and signed strength comparison.
#include "../player_check_strength_4f3180.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct test_modifier test_modifier;
typedef int32_t (*player_check_strength_fn)(nox_object_t*, nox_object_t*);

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	uint16_t type_index;
	int32_t strength;
};

struct test_modifier {
	uintptr_t marker;
	uint16_t required_strength;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"strength and result must remain exact 32-bit integers");
_Static_assert(sizeof(uint16_t) == 2,
	"type and required strength must remain exact 16-bit integers");
_Static_assert(
	_Generic(&nox_xxx_playerCheckStrength_4F3180,
		player_check_strength_fn: 1, default: 0),
	"strength eligibility must retain two native pointers and int32_t result");

static nox_object_t* expected_player;
static nox_object_t* live_item;
static test_modifier* weapon_definition;
static test_modifier* armor_definition;
static unsigned int sequence;
static int failure_line;
static int strength_calls;
static int weapon_calls;
static int armor_calls;
static int required_calls;

static int32_t get_strength(nox_object_t* player) {
	sequence = sequence * 10U + 1U;
	++strength_calls;
	if (player != expected_player)
		failure_line = __LINE__;
	if (live_item != NULL) {
		live_item->object_class = UINT32_C(0x02000000);
		live_item->type_index = UINT16_C(0xbeef);
	}
	return player->strength;
}

static test_modifier* find_weapon(uint16_t type_index) {
	sequence = sequence * 10U + 2U;
	++weapon_calls;
	if (type_index != UINT16_C(0x1234))
		failure_line = __LINE__;
	return weapon_definition;
}

static test_modifier* find_armor(uint16_t type_index) {
	sequence = sequence * 10U + 3U;
	++armor_calls;
	if (type_index != UINT16_C(0xbeef))
		failure_line = __LINE__;
	return armor_definition;
}

static uint16_t get_required(test_modifier* definition) {
	sequence = sequence * 10U + 4U;
	++required_calls;
	return definition->required_strength;
}

int32_t nox_xxx_playerCheckStrength_4F3180(
		nox_object_t* player,
		nox_object_t* item) {
	if (((uint8_t)player->object_class & UINT8_C(0x04)) == 0)
		return INT32_C(0);

	int32_t const strength = get_strength(player);
	uint32_t const item_class = item->object_class;
	uint16_t const type_index = item->type_index;
	test_modifier* const definition =
		(item_class & UINT32_C(0x02000000)) != 0
			? find_armor(type_index)
			: find_weapon(type_index);
	if (definition == NULL)
		return INT32_C(0);
	return strength >= (int32_t)get_required(definition)
		? INT32_C(1)
		: INT32_C(0);
}

static void reset_counters(void) {
	sequence = 0U;
	failure_line = 0;
	strength_calls = 0;
	weapon_calls = 0;
	armor_calls = 0;
	required_calls = 0;
}

int main(void) {
	nox_object_t player = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = UINT32_C(0x80000004),
		.strength = INT32_C(42),
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.object_class = UINT32_C(0x00001000),
		.type_index = UINT16_C(0x1234),
	};
	test_modifier weapon = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
		.required_strength = UINT16_C(42),
	};
	test_modifier armor = {
		.marker = UINTPTR_MAX - (uintptr_t)4,
		.required_strength = UINT16_C(42),
	};
	player_check_strength_fn const check =
		nox_xxx_playerCheckStrength_4F3180;

	expected_player = &player;
	weapon_definition = &weapon;
	armor_definition = &armor;
	live_item = &item;
	reset_counters();
	assert(check(&player, &item) == INT32_C(1));
	assert(failure_line == 0);
	assert(sequence == 134U);
	assert(strength_calls == 1 && weapon_calls == 0 && armor_calls == 1);
	assert(required_calls == 1);

	// A non-player returns after the class-byte load. NULL is safe here only
	// because the original never reaches the strength service or item fields.
	player.object_class = UINT32_C(0x00000400);
	live_item = NULL;
	reset_counters();
	assert(check(&player, NULL) == INT32_C(0));
	assert(sequence == 0U && strength_calls == 0);
	assert(weapon_calls == 0 && armor_calls == 0 && required_calls == 0);

	// Restore Player status, select the weapon path without mutation, and prove
	// a missing definition stops before the required-strength read.
	player.object_class = UINT32_C(0x00000004);
	player.strength = INT32_MAX;
	item.object_class = UINT32_C(0);
	item.type_index = UINT16_C(0x1234);
	live_item = NULL;
	weapon_definition = NULL;
	reset_counters();
	assert(check(&player, &item) == INT32_C(0));
	assert(sequence == 12U && weapon_calls == 1 && required_calls == 0);

	// The original comparison is signed int32 versus a zero-extended uint16.
	weapon_definition = &weapon;
	weapon.required_strength = UINT16_C(0);
	player.strength = INT32_MIN;
	reset_counters();
	assert(check(&player, &item) == INT32_C(0));
	assert(sequence == 124U);
	player.strength = INT32_C(0);
	reset_counters();
	assert(check(&player, &item) == INT32_C(1));
	assert(sequence == 124U);

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&player > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&weapon > UINT32_MAX);
		assert((uintptr_t)&armor > UINT32_MAX);
	}
	return 0;
}
