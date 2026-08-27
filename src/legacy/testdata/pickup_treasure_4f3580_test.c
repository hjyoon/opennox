// Freestanding C11 matrix companion for GAME.EXE 004F3580. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, wrapping counters, and solo/team completion paths.
#include "../pickup_treasure_4f3580.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_treasure_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

typedef struct player_t {
	uint32_t deaths;
	uint32_t count;
	uint32_t maximum;
	int32_t score;
} player_t;

typedef struct update_t {
	player_t* player;
} update_t;

typedef struct team_t {
	uint8_t id;
} team_t;

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	uint8_t team_id;
	update_t* update;
	nox_object_t* next;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4 && sizeof(uint32_t) == 4,
	"callback scalars, result, and counters must remain exact 32-bit values");
_Static_assert(
	_Generic(&nox_xxx_pickupTreasure_4F3580,
		pickup_treasure_fn: 1, default: 0),
	"TreasurePickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[96];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static nox_object_t* first_player;
static team_t* resolved_team;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t default_result;
static int32_t game_flag_result;
static uint32_t maximum_values[2];
static size_t maximum_call;
static uint32_t completion_flags;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	trace_count = 0U;
	expected_owner = owner;
	expected_item = item;
	first_player = NULL;
	resolved_team = NULL;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	default_result = INT32_C(0);
	game_flag_result = INT32_C(0);
	maximum_values[0] = UINT32_C(0);
	maximum_values[1] = UINT32_C(0);
	maximum_call = 0U;
	completion_flags = UINT32_C(0);
}

static int32_t load_arg4(int32_t value) {
	event(1U);
	return value;
}

static int32_t load_arg3(int32_t value) {
	event(2U);
	return value;
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

static uint8_t load_class_low(nox_object_t* owner) {
	event(4U);
	return (uint8_t)owner->object_class;
}

static update_t* load_update(nox_object_t* object) {
	event(5U);
	return object->update;
}

static int32_t game_flag(uint32_t flag) {
	event(6U);
	assert(flag == UINT32_C(0x40));
	return game_flag_result;
}

static void audio(uint32_t id, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(7U);
	assert(id == UINT32_C(307));
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
}

static player_t* load_player(update_t* update) {
	event(8U);
	return update->player;
}

static uint32_t load_count(player_t* player) {
	event(9U);
	return player->count;
}

static void store_count(player_t* player, uint32_t value) {
	event(10U);
	player->count = value;
}

static uint32_t treasure_maximum(void) {
	uint32_t value;
	event(11U);
	assert(maximum_call < 2U);
	value = maximum_values[maximum_call];
	maximum_call++;
	return value;
}

static void store_maximum(player_t* player, uint32_t value) {
	event(12U);
	player->maximum = value;
}

static void report(nox_object_t* owner) {
	event(13U);
	assert(owner == expected_owner);
}

static int32_t has_team(nox_object_t* owner) {
	event(14U);
	return owner->team_id != UINT8_C(0) ? INT32_C(1) : INT32_C(0);
}

static uint8_t load_object_team(nox_object_t* owner) {
	event(15U);
	return owner->team_id;
}

static team_t* find_team(uint8_t id) {
	event(16U);
	if (resolved_team != NULL)
		assert(id == resolved_team->id);
	return resolved_team;
}

static uint8_t load_team_id(team_t* team) {
	event(17U);
	return team->id;
}

static int32_t team_contains(nox_object_t* object, uint8_t id) {
	event(18U);
	return object->team_id != UINT8_C(0) && object->team_id == id
		? INT32_C(1)
		: INT32_C(0);
}

static nox_object_t* get_first_player(void) {
	event(19U);
	return first_player;
}

static nox_object_t* get_next_player(nox_object_t* current) {
	event(20U);
	return current->next;
}

static void set_game_flags(uint32_t flags) {
	event(21U);
	assert(flags == UINT32_C(0x08));
	completion_flags |= flags;
}

static void change_score(nox_object_t* owner, int32_t value) {
	event(22U);
	assert(owner == expected_owner);
	owner->update->player->score += value;
}

static void report_lesson(nox_object_t* object) {
	event(23U);
	assert(object != NULL);
}

static void increment_deaths(nox_object_t* object) {
	event(24U);
	object->update->player->deaths++;
}

int32_t nox_xxx_pickupTreasure_4F3580(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	update_t* update;
	player_t* player;
	uint32_t maximum;
	uint32_t total;
	nox_object_t* current;
	team_t* team;
	uint8_t team_id;

	arg4 = load_arg4(arg4);
	arg3 = load_arg3(arg3);
	if (default_pickup(owner, item, arg3, arg4) == INT32_C(0))
		return INT32_C(0);
	if ((load_class_low(owner) & UINT8_C(0x04)) == UINT8_C(0))
		return INT32_C(1);
	update = load_update(owner);
	if (game_flag(UINT32_C(0x40)) == INT32_C(0))
		return INT32_C(1);

	audio(UINT32_C(307), owner, INT32_C(0), UINT32_C(0));
	player = load_player(update);
	store_count(player, load_count(player) + UINT32_C(1));
	maximum = treasure_maximum();
	player = load_player(update);
	store_maximum(player, maximum);
	report(owner);

	if (has_team(owner) == INT32_C(0)) {
		maximum = treasure_maximum();
		player = load_player(update);
		if (load_count(player) == maximum) {
			set_game_flags(UINT32_C(0x08));
			change_score(owner, INT32_C(1));
			report_lesson(owner);
			for (current = get_first_player(); current != NULL;
				 current = get_next_player(current)) {
				if (current != owner) {
					increment_deaths(current);
					report_lesson(current);
				}
			}
		}
		return INT32_C(1);
	}

	team_id = load_object_team(owner);
	team = find_team(team_id);
	if (team == NULL)
		return INT32_C(1);
	total = UINT32_C(0);
	for (current = get_first_player(); current != NULL;
		 current = get_next_player(current)) {
		team_id = load_team_id(team);
		if (team_contains(current, team_id) != INT32_C(0)) {
			update_t* current_update = load_update(current);
			player_t* current_player = load_player(current_update);
			total += load_count(current_player);
		}
	}
	maximum = treasure_maximum();
	if (total == maximum)
		set_game_flags(UINT32_C(0x08));
	return INT32_C(1);
}

int main(void) {
	player_t owner_player = {0};
	player_t other_player = {0};
	player_t first_team_player = {0};
	player_t last_team_player = {0};
	update_t owner_update = {.player = &owner_player};
	update_t other_update = {.player = &other_player};
	update_t first_team_update = {.player = &first_team_player};
	update_t last_team_update = {.player = &last_team_player};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = UINT32_C(0x80000004),
		.update = &owner_update,
	};
	nox_object_t other = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.update = &other_update,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
	};
	nox_object_t first_team = {
		.marker = UINTPTR_MAX - (uintptr_t)4,
		.team_id = UINT8_C(7),
		.update = &first_team_update,
	};
	nox_object_t outsider = {
		.marker = UINTPTR_MAX - (uintptr_t)5,
		.team_id = UINT8_C(9),
	};
	nox_object_t last_team = {
		.marker = UINTPTR_MAX - (uintptr_t)6,
		.team_id = UINT8_C(7),
		.update = &last_team_update,
	};
	team_t team = {.id = UINT8_C(7)};

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}

	reset_case(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	assert(nox_xxx_pickupTreasure_4F3580(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_C(0));
	assert(trace_count == 3U);
	assert(trace[0] == 1U && trace[1] == 2U && trace[2] == 3U);

	reset_case(&owner, &item);
	expected_arg3 = INT32_C(-3);
	expected_arg4 = INT32_C(-4);
	default_result = INT32_MIN;
	game_flag_result = INT32_MIN;
	owner_player.count = UINT32_C(4);
	maximum_values[0] = UINT32_C(5);
	maximum_values[1] = UINT32_C(5);
	owner.next = &other;
	other.next = NULL;
	first_player = &owner;
	assert(nox_xxx_pickupTreasure_4F3580(
		&owner, &item, INT32_C(-3), INT32_C(-4)) == INT32_C(1));
	assert(owner_player.count == UINT32_C(5));
	assert(owner_player.maximum == UINT32_C(5));
	assert(owner_player.score == INT32_C(1));
	assert(other_player.deaths == UINT32_C(1));
	assert(completion_flags == UINT32_C(0x08));
	assert(maximum_call == 2U);

	reset_case(&owner, &item);
	default_result = INT32_C(1);
	game_flag_result = INT32_C(1);
	owner.team_id = UINT8_C(7);
	owner_player.count = UINT32_C(0);
	maximum_values[0] = UINT32_C(11);
	maximum_values[1] = UINT32_C(1);
	first_team_player.count = UINT32_MAX;
	last_team_player.count = UINT32_C(2);
	first_team.next = &outsider;
	outsider.next = &last_team;
	last_team.next = NULL;
	first_player = &first_team;
	resolved_team = &team;
	assert(nox_xxx_pickupTreasure_4F3580(
		&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(1));
	assert(owner_player.count == UINT32_C(1));
	assert(owner_player.maximum == UINT32_C(11));
	assert(completion_flags == UINT32_C(0x08));
	assert(maximum_call == 2U);

	return 0;
}
