// Freestanding C11 matrix companion for GAME.EXE 004F3D50 and 005367B0.
// Production uses the Go restoration; this file independently locks the
// callback ABI, native pointer width, signed result, ordered 51-row storage,
// first-duplicate lookup, zero-sound playback, and 50-row parser limit.
#include "../aud_event_pickup_4f3d50.h"

#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

enum {
	ROW_CAPACITY = 50,
	ROW_STORAGE = 51,
};

#define TYPE_SENTINEL UINT16_C(0xffff)

typedef int32_t (*aud_event_pickup_fn)(
	nox_object_t*, nox_object_t*, int32_t, int32_t);

struct nox_object_t {
	uintptr_t legacy_pointer;
	uint16_t type_ind;
};

typedef struct sound_row {
	uint16_t type_ind;
	uint16_t sound;
} sound_row;

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4 && sizeof(uint32_t) == 4,
	"callback scalars must remain exact 32-bit integers");
_Static_assert(offsetof(nox_object_t, type_ind) == sizeof(void*),
	"TypeInd must follow the native legacy pointer");
_Static_assert(sizeof(sound_row) == 4, "sound rows must remain four bytes");
_Static_assert(offsetof(sound_row, type_ind) == 0, "row type offset");
_Static_assert(offsetof(sound_row, sound) == 2, "row sound offset");
_Static_assert(
	_Generic(&nox_objectPickupAudEvent_4F3D50,
		aud_event_pickup_fn: 1, default: 0),
	"AudEventPickup must retain two native pointers, two int32_t values, and int32_t result");

static sound_row rows[ROW_STORAGE];
static uint32_t initialized;
static nox_object_t* expected_owner;
static nox_object_t* expected_item;
static int32_t expected_arg3;
static int32_t expected_arg4;
static int32_t default_result;
static unsigned int default_calls;
static unsigned int audio_calls;
static uint32_t last_sound;

static void reset_runtime(nox_object_t* owner, nox_object_t* item) {
	expected_owner = owner;
	expected_item = item;
	expected_arg3 = INT32_C(0);
	expected_arg4 = INT32_C(0);
	default_result = INT32_C(0);
	default_calls = 0U;
	audio_calls = 0U;
	last_sound = UINT32_C(0xffffffff);
}

static void reset_table(void) {
	initialized = UINT32_C(1);
	for (size_t row = 0; row < ROW_STORAGE; ++row) {
		rows[row].type_ind = TYPE_SENTINEL;
		rows[row].sound = UINT16_C(0);
	}
}

static int32_t default_pickup(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	++default_calls;
	assert(owner == expected_owner);
	assert(item == expected_item);
	assert(arg3 == expected_arg3);
	assert(arg4 == expected_arg4);
	return default_result;
}

static void audio(uint32_t sound, nox_object_t* owner, int32_t kind, uint32_t code) {
	++audio_calls;
	last_sound = sound;
	assert(owner == expected_owner);
	assert(kind == INT32_C(0));
	assert(code == UINT32_C(0));
}

int32_t nox_objectPickupAudEvent_4F3D50(
		nox_object_t* owner,
		nox_object_t* item,
		int32_t arg3,
		int32_t arg4) {
	if (owner == NULL)
		return INT32_C(0);
	if (item == NULL)
		return INT32_C(0);

	int32_t loaded_arg4 = arg4;
	int32_t loaded_arg3 = arg3;
	int32_t result = default_pickup(owner, item, loaded_arg3, loaded_arg4);
	if (result == INT32_C(0))
		return result;

	uint16_t row_type = rows[0].type_ind;
	if (row_type == TYPE_SENTINEL)
		return result;
	uint16_t type_ind = item->type_ind;
	for (size_t row = 0;; ++row) {
		if (row_type == type_ind) {
			audio((uint32_t)rows[row].sound, owner, INT32_C(0), UINT32_C(0));
			return result;
		}
		row_type = rows[row + 1U].type_ind;
		if (row_type == TYPE_SENTINEL)
			return result;
	}
}

static uint16_t resolve_sound(const char* token) {
	if (strcmp(token, "ApplePickup") == 0)
		return UINT16_C(101);
	if (strcmp(token, "PotionPickup") == 0)
		return UINT16_C(202);
	return UINT16_C(0);
}

static int parse_aud_event_pickup(uint16_t type_ind, const char* token) {
	if (initialized == UINT32_C(0)) {
		for (size_t row = 0; row < ROW_STORAGE; ++row) {
			rows[row].type_ind = TYPE_SENTINEL;
			rows[row].sound = UINT16_C(0);
		}
		initialized = UINT32_C(1);
	}

	size_t row = 0U;
	for (;;) {
		if (rows[row].type_ind == TYPE_SENTINEL)
			break;
		++row;
		if (row >= ROW_CAPACITY)
			return 0;
	}
	if (token == NULL || token[0] == '\0')
		return 0;
	uint16_t sound = resolve_sound(token);
	if (sound == UINT16_C(0))
		return 0;

	rows[row].sound = sound;
	rows[row].type_ind = type_ind;
	return 1;
}

int main(void) {
	nox_object_t owner = {
		.legacy_pointer = UINTPTR_MAX - (uintptr_t)1,
		.type_ind = UINT16_C(1),
	};
	nox_object_t item = {
		.legacy_pointer = UINTPTR_MAX - (uintptr_t)2,
		.type_ind = UINT16_C(44),
	};
	if (sizeof(void*) == 8) {
		assert((uintptr_t)&owner > UINT32_MAX);
		assert((uintptr_t)&item > UINT32_MAX);
	}

	reset_table();
	reset_runtime(&owner, &item);
	assert(nox_objectPickupAudEvent_4F3D50(
		NULL, &item, INT32_C(1), INT32_C(2)) == INT32_C(0));
	assert(nox_objectPickupAudEvent_4F3D50(
		&owner, NULL, INT32_C(1), INT32_C(2)) == INT32_C(0));
	assert(default_calls == 0U && audio_calls == 0U);

	reset_runtime(&owner, &item);
	expected_arg3 = INT32_MIN;
	expected_arg4 = INT32_MAX;
	default_result = INT32_MIN;
	assert(nox_objectPickupAudEvent_4F3D50(
		&owner, &item, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(default_calls == 1U && audio_calls == 0U);

	reset_table();
	rows[0] = (sound_row){.type_ind = UINT16_C(44), .sound = UINT16_C(0)};
	rows[1] = (sound_row){.type_ind = UINT16_C(44), .sound = UINT16_C(202)};
	reset_runtime(&owner, &item);
	expected_arg3 = INT32_C(-17);
	expected_arg4 = INT32_C(-23);
	default_result = INT32_C(-91);
	assert(nox_objectPickupAudEvent_4F3D50(
		&owner, &item, INT32_C(-17), INT32_C(-23)) == INT32_C(-91));
	assert(default_calls == 1U && audio_calls == 1U);
	assert(last_sound == UINT32_C(0));

	initialized = UINT32_C(0);
	assert(parse_aud_event_pickup(UINT16_C(44), "ApplePickup") == 1);
	assert(parse_aud_event_pickup(UINT16_C(44), "PotionPickup") == 1);
	assert(rows[0].type_ind == UINT16_C(44) && rows[0].sound == UINT16_C(101));
	assert(rows[1].type_ind == UINT16_C(44) && rows[1].sound == UINT16_C(202));
	assert(rows[2].type_ind == TYPE_SENTINEL);

	reset_runtime(&owner, &item);
	expected_arg3 = INT32_C(7);
	expected_arg4 = INT32_C(11);
	default_result = INT32_MAX;
	assert(nox_objectPickupAudEvent_4F3D50(
		&owner, &item, INT32_C(7), INT32_C(11)) == INT32_MAX);
	assert(default_calls == 1U && audio_calls == 1U);
	assert(last_sound == UINT32_C(101));

	initialized = UINT32_C(0);
	assert(parse_aud_event_pickup(UINT16_C(1), NULL) == 0);
	assert(initialized == UINT32_C(1));
	assert(rows[0].type_ind == TYPE_SENTINEL);
	assert(parse_aud_event_pickup(UINT16_C(2), "") == 0);
	assert(parse_aud_event_pickup(UINT16_C(3), "Unknown") == 0);
	assert(rows[0].type_ind == TYPE_SENTINEL);

	for (size_t row = 0; row < ROW_CAPACITY; ++row)
		assert(parse_aud_event_pickup((uint16_t)(row + 1U), "ApplePickup") == 1);
	assert(parse_aud_event_pickup(UINT16_C(99), "PotionPickup") == 0);
	assert(rows[ROW_CAPACITY].type_ind == TYPE_SENTINEL);

	return 0;
}
