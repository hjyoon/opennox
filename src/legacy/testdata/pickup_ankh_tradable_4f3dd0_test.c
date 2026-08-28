// Freestanding C11 matrix companion for GAME.EXE 004F3DD0. Production uses
// the Go restoration; this file independently locks the exact four-argument
// ABI, native pointer width, low class-byte test, cached UpdateData pointer,
// wrapping ExtraLives increment, late item load, deletion, and sound 1004.
#include "../pickup_ankh_tradable_4f3dd0.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef int32_t (*pickup_ankh_tradable_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

typedef struct player_update_t {
	uint32_t extra_lives;
} player_update_t;

struct nox_object_t {
	uintptr_t marker;
	uint32_t object_class;
	player_update_t* update;
};

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(uint32_t) == 4 && sizeof(int32_t) == 4,
	"object fields and callback values must remain exact 32-bit integers");
_Static_assert(offsetof(nox_object_t, object_class) == sizeof(uintptr_t),
	"class must follow the native-width legacy pointer");
_Static_assert(sizeof(((nox_object_t*)0)->update) == sizeof(void*),
	"UpdateData must remain a native pointer");
_Static_assert(sizeof(((player_update_t*)0)->extra_lives) == 4,
	"ExtraLives must remain uint32_t");
_Static_assert(
	_Generic(&nox_xxx_pickupAnkhTradable_4F3DD0,
		pickup_ankh_tradable_fn: 1, default: 0),
	"AnkhTradablePickup must retain two native pointers, two int32_t values, and int32_t result");

static unsigned int trace[16];
static size_t trace_count;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static nox_object_t* current_owner;
static nox_object_t* replacement_owner;
static player_update_t* replacement_update_on_load;
static unsigned int delete_calls;
static unsigned int audio_calls;

static void event(unsigned int value) {
	assert(trace_count < sizeof(trace) / sizeof(trace[0]));
	trace[trace_count++] = value;
}

static void reset_case(nox_object_t* owner, nox_object_t* item) {
	trace_count = 0U;
	expected_owner = owner;
	expected_item = item;
	current_owner = owner;
	replacement_owner = NULL;
	replacement_update_on_load = NULL;
	delete_calls = 0U;
	audio_calls = 0U;
}

static uint8_t load_class_low(nox_object_t* owner) {
	event(1U);
	return (uint8_t)owner->object_class;
}

static player_update_t* load_update(nox_object_t* owner) {
	event(2U);
	return owner->update;
}

static uint32_t load_extra_lives(player_update_t* update) {
	event(3U);
	uint32_t value = update->extra_lives;
	if (replacement_update_on_load != NULL)
		expected_owner->update = replacement_update_on_load;
	return value;
}

static void store_extra_lives(player_update_t* update, uint32_t value) {
	event(4U);
	update->extra_lives = value;
}

static nox_object_t* load_item_arg(nox_object_t* item) {
	event(5U);
	return item;
}

static void delayed_delete(nox_object_t* item) {
	event(6U);
	++delete_calls;
	assert(item == expected_item);
	if (replacement_owner != NULL)
		current_owner = replacement_owner;
}

static void audio(uint32_t sound, nox_object_t* owner, int32_t kind, uint32_t code) {
	event(7U);
	++audio_calls;
	assert(sound == UINT32_C(1004));
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
	if (replacement_owner != NULL)
		assert(current_owner == replacement_owner);
}

int32_t nox_xxx_pickupAnkhTradable_4F3DD0(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	(void)arg3;
	(void)arg4;
	if ((load_class_low(owner) & UINT8_C(0x04)) == UINT8_C(0))
		return INT32_C(0);

	player_update_t* update = load_update(owner);
	uint32_t extra_lives = load_extra_lives(update);
	store_extra_lives(update, extra_lives + UINT32_C(1));
	nox_object_t* loaded_item = load_item_arg(item);
	delayed_delete(loaded_item);
	audio(UINT32_C(1004), owner, INT32_C(0), UINT32_C(0));
	return INT32_C(1);
}

static void assert_trace(const unsigned int* want, size_t count) {
	assert(trace_count == count);
	for (size_t index = 0; index < count; ++index)
		assert(trace[index] == want[index]);
}

int main(void) {
	player_update_t entry_update = {.extra_lives = UINT32_C(9)};
	player_update_t replacement_update = {.extra_lives = UINT32_C(77)};
	nox_object_t owner = {
		.marker = UINTPTR_MAX - (uintptr_t)1,
		.object_class = UINT32_C(0xa50000a4),
		.update = &entry_update,
	};
	nox_object_t item = {.marker = UINTPTR_MAX - (uintptr_t)2};
	nox_object_t owner_replacement = {.marker = UINTPTR_MAX - (uintptr_t)3};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
		assert((uintptr_t)&entry_update > UINT32_MAX);
	}

	reset_case(&owner, &item);
	assert(nox_xxx_pickupAnkhTradable_4F3DD0(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_C(1));
	assert(entry_update.extra_lives == UINT32_C(10));
	assert(delete_calls == 1U && audio_calls == 1U);
	{
		const unsigned int want[] = {1U, 2U, 3U, 4U, 5U, 6U, 7U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	owner.object_class = UINT32_C(0xfffffffb);
	owner.update = NULL;
	reset_case(&owner, NULL);
	assert(nox_xxx_pickupAnkhTradable_4F3DD0(
		&owner, NULL, INT32_MIN, INT32_MAX) == INT32_C(0));
	assert(delete_calls == 0U && audio_calls == 0U);
	{
		const unsigned int want[] = {1U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	entry_update.extra_lives = UINT32_MAX;
	replacement_update.extra_lives = UINT32_C(77);
	owner.object_class = UINT32_C(0x04);
	owner.update = &entry_update;
	reset_case(&owner, &item);
	replacement_update_on_load = &replacement_update;
	replacement_owner = &owner_replacement;
	assert(nox_xxx_pickupAnkhTradable_4F3DD0(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(1));
	assert(entry_update.extra_lives == UINT32_C(0));
	assert(replacement_update.extra_lives == UINT32_C(77));
	assert(owner.update == &replacement_update);
	assert(current_owner == &owner_replacement);
	{
		const unsigned int want[] = {1U, 2U, 3U, 4U, 5U, 6U, 7U};
		assert_trace(want, sizeof(want) / sizeof(want[0]));
	}

	entry_update.extra_lives = UINT32_C(4);
	owner.update = &entry_update;
	reset_case(&owner, NULL);
	assert(nox_xxx_pickupAnkhTradable_4F3DD0(
		&owner, NULL, INT32_C(0), INT32_C(0)) == INT32_C(1));
	assert(entry_update.extra_lives == UINT32_C(5));
	assert(delete_calls == 1U && audio_calls == 1U);

	return 0;
}
