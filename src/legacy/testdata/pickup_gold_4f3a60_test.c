// Freestanding C11 matrix companion for GAME.EXE 004F3A60 and its 004FA590
// gold-add dependency. Production uses the Go restoration; this file
// independently locks the exact four-argument ABI, native pointer width,
// cached GoldInitData behavior, 32-bit gold wrapping, Player reload, localized
// message amount, and full non-Player int32 result.
#include "../pickup_gold_4f3a60.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

typedef int32_t (*pickup_gold_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

typedef struct gold_data_t {
	uint32_t amount;
} gold_data_t;

typedef struct player_t {
	uint32_t gold;
	uint32_t protection_token;
} player_t;

typedef struct player_update_t {
	player_t* player;
} player_update_t;

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	gold_data_t* init_data;
	player_update_t* update_data;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(uint32_t) == 4 && sizeof(int32_t) == 4,
	"gold and callback values must remain exact 32-bit integers");
_Static_assert(sizeof(gold_data_t) == 4,
	"GoldInitData must remain one uint32 amount");
_Static_assert(
	_Generic(&nox_xxx_pickupGold_4F3A60,
		pickup_gold_fn: 1, default: 0),
	"GoldPickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[32];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t default_result;
static gold_data_t* cached_data;
static gold_data_t* replacement_data;
static player_t* replacement_player;

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
	cached_data = NULL;
	replacement_data = NULL;
	replacement_player = NULL;
}

static uint8_t load_class_low(nox_object_t* owner) {
	event(1U);
	return (uint8_t)owner->object_class;
}

static gold_data_t* load_init_data(nox_object_t* item) {
	event(2U);
	return item->init_data;
}

static uint32_t load_amount(gold_data_t* data) {
	event(3U);
	return data->amount;
}

static player_update_t* load_update(nox_object_t* owner) {
	event(4U);
	return owner->update_data;
}

static player_t* load_player(player_update_t* update) {
	event(5U);
	return update->player;
}

static uint32_t load_gold(player_t* player) {
	event(6U);
	return player->gold;
}

static void store_gold(player_update_t* update, player_t* player, uint32_t value) {
	event(7U);
	player->gold = value;
	if (replacement_player != NULL)
		update->player = replacement_player;
}

static uint32_t load_token(player_t* player) {
	event(8U);
	return player->protection_token;
}

static void protect_gold(uint32_t token, int32_t delta) {
	event(9U);
	assert(token == UINT32_C(0xfedcba98));
	assert((uint32_t)delta == UINT32_C(0x80000001));
}

static void add_gold(nox_object_t* owner, uint32_t amount) {
	player_update_t* update = load_update(owner);
	player_t* player = load_player(update);
	uint32_t gold = load_gold(player);
	store_gold(update, player, gold + amount);
	player = load_player(update);
	protect_gold(load_token(player), (int32_t)amount);
}

static void delayed_delete(nox_object_t* item) {
	event(10U);
	assert(item == expected_item);
	if (cached_data != NULL)
		cached_data->amount = UINT32_C(0xfedcba98);
	if (replacement_data != NULL) {
		replacement_data->amount = UINT32_C(0x22222222);
		item->init_data = replacement_data;
	}
}

static const char* load_string(const char* key, const char* path, int line) {
	event(11U);
	assert(strcmp(key, "GoldPickup") == 0);
	assert(strcmp(path, "C:\\NoxPost\\src\\Server\\Object\\pickdrop\\pickup.c") == 0);
	assert(line == 709);
	return "localized";
}

static void send_line_message(
		nox_object_t* owner, const char* message, uint32_t amount) {
	event(12U);
	assert(owner == expected_owner);
	assert(strcmp(message, "localized") == 0);
	assert(amount == UINT32_C(0xfedcba98));
}

static void audio(uint32_t id, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(13U);
	assert(id == UINT32_C(307));
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
}

static int32_t load_arg4(int32_t value) {
	event(14U);
	return value;
}

static int32_t load_arg3(int32_t value) {
	event(15U);
	return value;
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	event(16U);
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

int32_t nox_xxx_pickupGold_4F3A60(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	if ((load_class_low(owner) & UINT8_C(0x04)) != UINT8_C(0)) {
		gold_data_t* data = load_init_data(item);
		uint32_t amount = load_amount(data);
		add_gold(owner, amount);
		delayed_delete(item);
		amount = load_amount(data);
		const char* message = load_string(
			"GoldPickup",
			"C:\\NoxPost\\src\\Server\\Object\\pickdrop\\pickup.c",
			709);
		send_line_message(owner, message, amount);
		audio(UINT32_C(307), owner, INT32_C(0), UINT32_C(0));
		return INT32_C(1);
	}

	int32_t loaded_arg4 = load_arg4(arg4);
	int32_t loaded_arg3 = load_arg3(arg3);
	int32_t result = default_pickup(owner, item, loaded_arg3, loaded_arg4);
	if (result != INT32_C(0))
		audio(UINT32_C(307), owner, INT32_C(0), UINT32_C(0));
	return result;
}

int main(void) {
	player_t first = {
		.gold = UINT32_MAX,
		.protection_token = UINT32_C(0x11111111),
	};
	player_t second = {
		.gold = UINT32_C(0x22222222),
		.protection_token = UINT32_C(0xfedcba98),
	};
	player_update_t update = {.player = &first};
	gold_data_t original = {.amount = UINT32_C(0x80000001)};
	gold_data_t replacement = {.amount = UINT32_C(0x11111111)};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = UINT32_C(0x80000004),
		.update_data = &update,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.init_data = &original,
	};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&update > UINT32_MAX);
		assert((uintptr_t)&original > UINT32_MAX);
	}

	reset_case(&owner, &item);
	cached_data = &original;
	replacement_data = &replacement;
	replacement_player = &second;
	assert(nox_xxx_pickupGold_4F3A60(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_C(1));
	assert(first.gold == UINT32_C(0x80000000));
	assert(update.player == &second);
	assert(item.init_data == &replacement);
	assert(trace_count == 15U);
	{
		const unsigned int want[] = {
			1U, 2U, 3U, 4U, 5U, 6U, 7U, 5U, 8U, 9U,
			10U, 3U, 11U, 12U, 13U,
		};
		for (size_t i = 0; i < sizeof(want) / sizeof(want[0]); ++i)
			assert(trace[i] == want[i]);
	}

	owner.object_class = UINT32_C(0x80000000);
	item.init_data = NULL;
	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_result = INT32_MIN;
	assert(nox_xxx_pickupGold_4F3A60(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(trace_count == 5U);
	assert(trace[0] == 1U && trace[1] == 14U && trace[2] == 15U);
	assert(trace[3] == 16U && trace[4] == 13U);

	reset_case(&owner, NULL);
	expected_item = NULL;
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	assert(nox_xxx_pickupGold_4F3A60(
		&owner, NULL, INT32_C(-17), INT32_C(-23)) == INT32_C(0));
	assert(trace_count == 4U);
	assert(trace[0] == 1U && trace[1] == 14U && trace[2] == 15U && trace[3] == 16U);

	return 0;
}
