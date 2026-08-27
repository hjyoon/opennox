// Freestanding C11 matrix companion for GAME.EXE 004F31E0. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, team rejection, signed wrapping weight rules,
// food limits, and delete-before-insert ordering.
#include "../pickup_default_4f31e0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef struct test_player test_player;
typedef struct test_team test_team;
typedef struct test_update test_update;
typedef int32_t (*pickup_default_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

struct test_player {
	uintptr_t marker;
	uint8_t index;
};

struct test_update {
	uintptr_t marker;
	test_player* player;
};

struct test_team {
	uintptr_t marker;
	uint8_t id;
	uint8_t color;
};

struct nox_object_t {
	uintptr_t marker;
	uint16_t type_index;
	uint32_t object_class;
	uint8_t team_id;
	uint8_t weight;
	uint16_t capacity;
	nox_object_t* holder;
	nox_object_t* next;
	nox_object_t* first;
	test_update* update;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"callback scalars and result must remain exact 32-bit integers");
_Static_assert(sizeof(uint16_t) == 2,
	"type and capacity must remain exact 16-bit integers");
_Static_assert(
	_Generic(&nox_xxx_pickupDefault_4F31E0,
		pickup_default_fn: 1, default: 0),
	"default pickup must retain two native pointers, two int32_t values, and int32_t result");

static uint32_t game_mode;
static test_team* live_team;
static int32_t live_count;
static unsigned int trace[32];
static size_t trace_count;
static int failure_line;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_report;
static uint8_t informed_index;
static uint8_t informed_code;
static uint32_t informed_color;
static char const* primary_message;

static void event(unsigned int code) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = code;
}

static int32_t game_flags_check(uint32_t mask) {
	event(mask == UINT32_C(0x1000) ? 1U : 2U);
	return (game_mode & mask) == mask ? INT32_C(1) : INT32_C(0);
}

static test_team* find_team(uint8_t id) {
	event(3U);
	return live_team != NULL && live_team->id == id ? live_team : NULL;
}

static void inform_team(uint8_t index, uint8_t code, uint32_t color) {
	event(4U);
	informed_index = index;
	informed_code = code;
	informed_color = color;
}

static void send_primary(nox_object_t* owner, char const* message, uint8_t value) {
	event(5U);
	if (owner != expected_owner || value != UINT8_C(0))
		failure_line = __LINE__;
	primary_message = message;
}

static int32_t count_inventory(nox_object_t* owner, int32_t type_index) {
	event(6U);
	if (owner != expected_owner || type_index != (int32_t)expected_item->type_index)
		failure_line = __LINE__;
	return live_count;
}

static void delete_world_object(nox_object_t* item) {
	event(7U);
	if (item != expected_item)
		failure_line = __LINE__;
}

static void inventory_put(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t report) {
	event(8U);
	if (owner != expected_owner || item != expected_item ||
		report != expected_report || trace_count < 2U || trace[trace_count - 2U] != 7U)
		failure_line = __LINE__;
}

static int32_t wrapping_budget(uint16_t capacity, uint32_t inventory_weight) {
	return (int32_t)(((uint32_t)capacity << 1U) - inventory_weight);
}

static int food_below_limit(int32_t count, int32_t limit) {
	return (int32_t)((uint32_t)count - (uint32_t)limit) < INT32_C(0);
}

int32_t nox_xxx_pickupDefault_4F31E0(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t report,
		int32_t ignored) {
	(void)ignored;
	if (game_flags_check(UINT32_C(0x1000)) == 0 &&
		item->team_id != UINT8_C(0) &&
		(owner->team_id == UINT8_C(0) || owner->team_id != item->team_id)) {
		uint8_t const team_id = item->team_id;
		test_team* const team = find_team(team_id);
		if (team != NULL) {
			if (((uint8_t)owner->object_class & UINT8_C(0x04)) != 0) {
				test_update* const update = owner->update;
				uint8_t const color = team->color;
				test_player* const player = update->player;
				uint8_t const index = player->index;
				inform_team(index, UINT8_C(16), (uint32_t)color);
			}
			return INT32_C(0);
		}
	}

	if (item->holder != NULL)
		return INT32_C(0);
	uint16_t const capacity = owner->capacity;
	if (capacity == UINT16_C(0))
		return INT32_C(0);

	uint32_t inventory_weight = UINT32_C(0);
	for (nox_object_t* current = owner->first; current != NULL;) {
		uint8_t const weight = current->weight;
		current = current->next;
		inventory_weight += (uint32_t)weight;
	}
	uint8_t const item_weight = item->weight;
	if (wrapping_budget(capacity, inventory_weight) < (int32_t)item_weight) {
		send_primary(owner, "pickup.c:CarryingTooMuch", UINT8_C(0));
		return INT32_C(0);
	}

	if ((item->object_class & UINT32_C(0x10)) != 0) {
		uint16_t const type_index = item->type_index;
		int32_t const count = count_inventory(owner, (int32_t)type_index);
		int32_t limit = INT32_C(3);
		if (game_flags_check(UINT32_C(0x1800)) != 0)
			limit = INT32_C(9);
		if (!food_below_limit(count, limit)) {
			send_primary(owner, "pickup.c:MaxSameItem", UINT8_C(0));
			return INT32_C(0);
		}
	}

	delete_world_object(item);
	inventory_put(owner, item, report);
	return INT32_C(1);
}

static void reset_case(nox_object_t* owner, nox_object_t* item, int32_t report) {
	game_mode = UINT32_C(0);
	live_team = NULL;
	live_count = INT32_C(0);
	trace_count = 0U;
	failure_line = 0;
	expected_owner = owner;
	expected_item = item;
	expected_report = report;
	informed_index = UINT8_C(0);
	informed_code = UINT8_C(0);
	informed_color = UINT32_C(0);
	primary_message = NULL;
}

int main(void) {
	test_player player = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.index = UINT8_C(7),
	};
	test_update update = {
		.marker = UINTPTR_MAX - (uintptr_t)2,
		.player = &player,
	};
	test_team team = {
		.marker = UINTPTR_MAX - (uintptr_t)3,
		.id = UINT8_C(9),
		.color = UINT8_C(6),
	};
	nox_object_t second = {
		.marker = UINTPTR_MAX - (uintptr_t)4,
		.weight = UINT8_C(17),
	};
	nox_object_t first = {
		.marker = UINTPTR_MAX - (uintptr_t)5,
		.weight = UINT8_C(11),
		.next = &second,
	};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)6,
		.object_class = UINT32_C(0x80000004),
		.team_id = UINT8_C(3),
		.capacity = UINT16_C(50),
		.first = &first,
		.update = &update,
	};
	nox_object_t item = {
		.marker = UINTPTR_MAX - (uintptr_t)7,
		.type_index = UINT16_C(0x1234),
		.object_class = UINT32_C(0x80000010),
		.weight = UINT8_C(20),
	};
	pickup_default_fn const pickup = nox_xxx_pickupDefault_4F31E0;

	// Quest bypasses team filtering, raises the Food limit to nine, and still
	// forwards only report to inventory insertion. The fourth value is ignored.
	reset_case(&owner, &item, INT32_MIN);
	game_mode = UINT32_C(0x1800);
	live_count = INT32_C(8);
	assert(pickup(&owner, &item, INT32_MIN, INT32_MAX) == INT32_C(1));
	assert(failure_line == 0);
	assert(trace_count == 5U);
	assert(trace[0] == 1U && trace[1] == 6U && trace[2] == 2U);
	assert(trace[3] == 7U && trace[4] == 8U);

	// A live mismatched team rejects before any inventory read and informs the
	// Player with the resolved team's color.
	item.team_id = UINT8_C(9);
	reset_case(&owner, &item, INT32_C(1));
	live_team = &team;
	assert(pickup(&owner, &item, INT32_C(1), INT32_C(-1)) == INT32_C(0));
	assert(failure_line == 0);
	assert(trace_count == 3U && trace[0] == 1U && trace[1] == 3U && trace[2] == 4U);
	assert(informed_index == UINT8_C(7));
	assert(informed_code == UINT8_C(16));
	assert(informed_color == UINT32_C(6));

	// A stale team ID whose Team no longer resolves is allowed to continue.
	reset_case(&owner, &item, INT32_C(11));
	item.object_class = UINT32_C(0x80000000);
	assert(pickup(&owner, &item, INT32_C(11), INT32_C(12)) == INT32_C(1));
	assert(failure_line == 0);
	assert(trace_count == 4U && trace[0] == 1U && trace[1] == 3U);
	assert(trace[2] == 7U && trace[3] == 8U);

	// Weight rejection uses a signed int32_t view of a wrapping uint32_t budget.
	item.team_id = UINT8_C(0);
	item.weight = UINT8_C(73);
	owner.capacity = UINT16_C(50);
	reset_case(&owner, &item, INT32_C(0));
	assert(pickup(&owner, &item, INT32_C(0), INT32_C(0)) == INT32_C(0));
	assert(failure_line == 0 && trace_count == 2U);
	assert(trace[0] == 1U && trace[1] == 5U);
	assert(primary_message != NULL && primary_message[9] == 'C');
	assert(wrapping_budget(UINT16_MAX, UINT32_MAX) == INT32_C(0x1ffff));
	assert(wrapping_budget(UINT16_C(1), UINT32_C(3)) == INT32_C(-1));
	assert(!food_below_limit(INT32_MIN, INT32_C(3)));
	assert(!food_below_limit(INT32_MAX, INT32_C(9)));

	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&player > UINT32_MAX);
		assert((uintptr_t)&update > UINT32_MAX);
		assert((uintptr_t)&team > UINT32_MAX);
	}
	return 0;
}
