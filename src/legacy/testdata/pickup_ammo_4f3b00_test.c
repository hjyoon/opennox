// Freestanding C11 matrix companion for GAME.EXE 004F3B00. Production uses
// the Go restoration; this independently locks the four-argument ABI, native
// pointer width, four native modifier identities, exact three-byte AmmoUseData,
// uint8 charge wrapping, capacity boundary, cached update pointer, and full
// WeaponPickup fallback result.
#include "../pickup_ammo_4f3b00.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_ammo_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

typedef struct modifier_t {
	uintptr_t marker;
} modifier_t;

typedef struct modifier_init_t {
	modifier_t* modifiers[4];
	uint32_t field16;
} modifier_init_t;

typedef struct ammo_use_t {
	uint8_t charge0;
	uint8_t charge1;
	uint8_t field2;
} ammo_use_t;

typedef struct player_t {
	uint8_t player_ind;
} player_t;

typedef struct player_update_t {
	player_t* player;
} player_update_t;

struct nox_object_t {
	uintptr_t marker;
	uint16_t type_ind;
	uint32_t object_class;
	uint32_t equip_flags;
	nox_object_t* inventory_first;
	nox_object_t* inventory_next;
	modifier_init_t* init_data;
	ammo_use_t* use_data;
	player_update_t* update_data;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(uint8_t) == 1 && sizeof(uint16_t) == 2 &&
	sizeof(uint32_t) == 4 && sizeof(int32_t) == 4,
	"AmmoPickup fixed-width values changed size");
_Static_assert(sizeof(ammo_use_t) == 3 &&
	offsetof(ammo_use_t, charge0) == 0 &&
	offsetof(ammo_use_t, charge1) == 1 &&
	offsetof(ammo_use_t, field2) == 2,
	"AmmoUseData must remain the exact three-byte payload");
_Static_assert(sizeof(((modifier_init_t*)0)->modifiers) == 4 * sizeof(void*),
	"ModifierInitData must retain four native pointers");
_Static_assert(offsetof(modifier_init_t, field16) == 4 * sizeof(void*),
	"ModifierInitData trailing scalar must follow four native pointers");
_Static_assert(
	_Generic(&nox_xxx_pickupAmmo_4F3B00,
		pickup_ammo_fn: 1, default: 0),
	"AmmoPickup must retain two native pointers, two int32_t values, and int32_t result");

#define CLASS_PLAYER UINT32_C(0x00000004)
#define CLASS_WEAPON UINT32_C(0x01000000)
#define AMMO_MASK UINT32_C(0x00000082)
#define AMMO_CAPACITY UINT16_C(250)

static unsigned int trace[32];
static size_t trace_count;
static size_t modifier_reads;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static nox_object_t* expected_candidate;
static player_update_t* replacement_update;
static int mutate_item_flags;
static int mutate_owner_update;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t weapon_pickup_result;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	trace_count = 0U;
	modifier_reads = 0U;
	expected_owner = owner;
	expected_item = item;
	expected_candidate = NULL;
	replacement_update = NULL;
	mutate_item_flags = 0;
	mutate_owner_update = 0;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	weapon_pickup_result = INT32_C(0);
}

static uint32_t weapon_equip_flags(nox_object_t* object) {
	uint32_t const value = object == NULL ? UINT32_C(0) : object->equip_flags;
	if (object == expected_item) {
		event(1U);
		if (mutate_item_flags)
			object->equip_flags = UINT32_C(0);
	} else {
		event(2U);
		assert(object == expected_candidate);
		if (mutate_owner_update)
			expected_owner->update_data = replacement_update;
	}
	return value;
}

static modifier_t* load_modifier(modifier_init_t* data, size_t index) {
	++modifier_reads;
	return data->modifiers[index];
}

static int32_t load_arg4(int32_t value) {
	event(3U);
	return value;
}

static int32_t load_arg3(int32_t value) {
	event(4U);
	return value;
}

static int32_t weapon_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(5U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return weapon_pickup_result;
}

static void report_charges(
		uint8_t player_ind,
		nox_object_t* candidate,
		uint8_t charge1,
		uint8_t charge0) {
	event(6U);
	assert(player_ind == UINT8_C(0xfe));
	assert(candidate == expected_candidate);
	assert(charge1 == UINT8_C(4));
	assert(charge0 == UINT8_C(250));
}

static void delayed_delete(nox_object_t* item) {
	event(7U);
	assert(item == expected_item);
}

static void pickup_audio(nox_object_t* owner, nox_object_t* item) {
	event(8U);
	assert(owner == expected_owner);
	assert(item == expected_item);
}

static int32_t fallback(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	int32_t const loaded_arg4 = load_arg4(arg4);
	int32_t const loaded_arg3 = load_arg3(arg3);
	return weapon_pickup(owner, item, loaded_arg3, loaded_arg4);
}

int32_t nox_xxx_pickupAmmo_4F3B00(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	uint32_t const item_flags = weapon_equip_flags(item);
	if (((uint8_t)owner->object_class & UINT8_C(0x04)) == UINT8_C(0))
		return fallback(owner, item, arg3, arg4);

	player_update_t* const update = owner->update_data;
	if ((item_flags & AMMO_MASK) == UINT32_C(0))
		return fallback(owner, item, arg3, arg4);

	nox_object_t* candidate = owner->inventory_first;
	modifier_init_t* const item_init = item->init_data;
	ammo_use_t* const item_use = item->use_data;
	while (candidate != NULL) {
		if (candidate->type_ind == item->type_ind &&
			(candidate->object_class & CLASS_WEAPON) != UINT32_C(0) &&
			(weapon_equip_flags(candidate) & item_flags) != UINT32_C(0)) {
			modifier_init_t* const candidate_init = candidate->init_data;
			ammo_use_t* const candidate_use = candidate->use_data;
			int modifiers_equal = 1;
			for (size_t index = 0U; index < 4U; ++index) {
				modifier_t* const candidate_modifier =
					load_modifier(candidate_init, index);
				modifier_t* const item_modifier = load_modifier(item_init, index);
				if (candidate_modifier != item_modifier)
					modifiers_equal = 0;
			}
			if (modifiers_equal && candidate_use->field2 == UINT8_C(0) &&
				(uint16_t)candidate_use->charge0 + (uint16_t)item_use->charge0 <=
					AMMO_CAPACITY) {
				uint8_t const merged_charge1 =
					(uint8_t)(item_use->charge1 + candidate_use->charge1);
				candidate_use->charge1 = merged_charge1;
				uint8_t const merged_charge0 =
					(uint8_t)(candidate_use->charge0 + item_use->charge0);
				candidate_use->charge0 = merged_charge0;
				player_t* const player = update->player;
				report_charges(
					player->player_ind,
					candidate,
					merged_charge1,
					merged_charge0);
				delayed_delete(item);
				pickup_audio(owner, item);
				return INT32_C(1);
			}
		}
		candidate = candidate->inventory_next;
	}
	return fallback(owner, item, arg3, arg4);
}

int main(void) {
	modifier_t mod0 = {.marker = UINTPTR_MAX - (uintptr_t)1};
	modifier_t mod1 = {.marker = UINTPTR_MAX - (uintptr_t)2};
	modifier_t mod3 = {.marker = UINTPTR_MAX - (uintptr_t)3};
	modifier_init_t item_init = {
		.modifiers = {&mod0, &mod1, NULL, &mod3},
	};
	modifier_init_t candidate_init = {
		.modifiers = {&mod0, &mod1, NULL, &mod3},
	};
	ammo_use_t item_use = {
		.charge0 = UINT8_C(50),
		.charge1 = UINT8_C(250),
		.field2 = UINT8_MAX,
	};
	ammo_use_t candidate_use = {
		.charge0 = UINT8_C(200),
		.charge1 = UINT8_C(10),
	};
	player_t cached_player = {.player_ind = UINT8_C(0xfe)};
	player_t replacement_player = {.player_ind = UINT8_C(1)};
	player_update_t cached_update = {.player = &cached_player};
	player_update_t other_update = {.player = &replacement_player};
	nox_object_t candidate = {
		.marker = UINTPTR_MAX - (uintptr_t)4,
		.type_ind = UINT16_C(0x8001),
		.object_class = CLASS_WEAPON,
		.equip_flags = UINT32_C(0x80),
		.init_data = &candidate_init,
		.use_data = &candidate_use,
	};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)5,
		.object_class = UINT32_C(0x80000004),
		.inventory_first = &candidate,
		.update_data = &cached_update,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)6,
		.type_ind = UINT16_C(0x8001),
		.equip_flags = AMMO_MASK,
		.init_data = &item_init,
		.use_data = &item_use,
	};

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&candidate > UINT32_MAX);
		assert((uintptr_t)&item_init > UINT32_MAX);
		assert((uintptr_t)&candidate_use > UINT32_MAX);
	}

	reset_case(&owner, &item);
	expected_candidate = &candidate;
	replacement_update = &other_update;
	mutate_item_flags = 1;
	mutate_owner_update = 1;
	assert(nox_xxx_pickupAmmo_4F3B00(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_C(1));
	assert(item.equip_flags == UINT32_C(0));
	assert(owner.update_data == &other_update);
	assert(candidate_use.charge0 == UINT8_C(250));
	assert(candidate_use.charge1 == UINT8_C(4));
	assert(modifier_reads == 8U);
	assert(trace_count == 5U);
	assert(trace[0] == 1U && trace[1] == 2U && trace[2] == 6U);
	assert(trace[3] == 7U && trace[4] == 8U);

	owner.object_class = UINT32_C(0x80000000);
	item.equip_flags = AMMO_MASK;
	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	weapon_pickup_result = INT32_MIN;
	assert(nox_xxx_pickupAmmo_4F3B00(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(trace_count == 4U);
	assert(trace[0] == 1U && trace[1] == 3U && trace[2] == 4U && trace[3] == 5U);

	owner.object_class = CLASS_PLAYER;
	owner.update_data = &cached_update;
	owner.inventory_first = &candidate;
	item.equip_flags = AMMO_MASK;
	item_use.charge0 = UINT8_C(1);
	item_use.charge1 = UINT8_C(0);
	candidate_use.charge0 = UINT8_C(250);
	candidate_use.charge1 = UINT8_C(0);
	candidate_use.field2 = UINT8_C(0);
	reset_case(&owner, &item);
	expected_candidate = &candidate;
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	weapon_pickup_result = INT32_C(-99);
	assert(nox_xxx_pickupAmmo_4F3B00(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(-99));
	assert(candidate_use.charge0 == UINT8_C(250));
	assert(modifier_reads == 8U);
	assert(trace_count == 5U);
	assert(trace[0] == 1U && trace[1] == 2U && trace[2] == 3U);
	assert(trace[3] == 4U && trace[4] == 5U);

	modifier_t different = {.marker = (uintptr_t)7};
	candidate_init.modifiers[0] = &different;
	item_use.charge0 = UINT8_C(0);
	candidate_use.charge0 = UINT8_C(0);
	reset_case(&owner, &item);
	expected_candidate = &candidate;
	weapon_pickup_result = INT32_C(0);
	assert(nox_xxx_pickupAmmo_4F3B00(
		&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(0));
	assert(modifier_reads == 8U);

	return 0;
}
