#include <stdio.h>

// Suppress unrelated PE32 aggregate assertions, not the production secret-wall
// header or the native wall assertions extracted with the production layout.
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert
#include "../secret_wall.h"

// PRODUCTION_WALL_LAYOUT

static nox_wall_native_t* expected_wall;
static nox_secret_wall_t* expected_data;
static nox_secret_wall_t* next_data;
static nox_object_t* expected_unit;
static nox_wall_native_t* lookup_result;
static unsigned int sequence;
static unsigned int sound_calls;
static int failure;
static int mutate_tile;
static char sound_name[] = "native-wall-open";
static float2 wanted_pos;
static int2 lookup_grid = {37, 91};

#define CHECK(test) do { if (!(test)) { \
	fprintf(stderr, "line %d: %s (class=%#x flags=%u state=%u delay=%u sequence=%u)\n", \
		__LINE__, #test, expected_unit->obj_class, expected_data->flags, \
		expected_data->state, expected_data->open_delay, sequence); return 1; \
} } while (0)

static nox_wall_native_t* nox_server_getWallAtGrid_410580(int x, int y) {
	if (x != lookup_grid.field_0 || y != lookup_grid.field_4 || sequence) failure = 1;
	sequence = 1;
	// The original reads the wall flags and tile after wall lookup.
	if (mutate_tile) expected_wall->tile = 255;
	return lookup_result;
}

static char* nox_xxx_wallFindOpenSound_410EE0(unsigned int tile) {
	if (sequence != 1 || tile != expected_wall->tile ||
		expected_data->state != 4 || expected_data->open_delay != 0) failure = 1;
	sequence = 12;
	return sound_name;
}

static int nox_xxx_utilFindSound_40AF50(char* name) {
	if (sequence != 12 || name != sound_name) failure = 1;
	sequence = 123;
	return 0x12345678;
}

static void nox_xxx_audCreate_501A30(int id, float2* pos, int kind, int extra) {
	if (sequence != 123 || id != 0x12345678 || kind || extra ||
		memcmp(pos, &wanted_pos, sizeof(*pos)) || expected_data->state != 4 ||
		expected_data->open_delay) failure = 1;
	sequence = 1234;
	sound_calls++;
}

// PRODUCTION_BODY_548100

// Independently model the decoded LEA/SHL/SUB DWORD wrap, signed FILD, FADD
// 11.5, and FSTP binary32. In particular, this is not unsigned float scaling.
static float original_center(int32_t grid) {
	uint64_t product = (uint64_t)(uint32_t)grid * UINT64_C(23);
	int64_t signed_product = (int64_t)(product & UINT64_C(0xFFFFFFFF));
	if (signed_product > INT32_MAX) signed_product -= INT64_C(0x100000000);
	return (float)((double)signed_product + 11.5);
}

static void prepare(unsigned int wall_flags, unsigned int unit_class,
	unsigned int flags, unsigned int state, unsigned int delay, int32_t x, int32_t y) {
	memset(expected_wall, 0xA5, sizeof(*expected_wall));
	memset(expected_data, 0xA5, sizeof(*expected_data));
	memset(expected_unit, 0xA5, sizeof(*expected_unit));
	expected_wall->flags = (uint8_t)wall_flags;
	expected_wall->tile = 128;
	expected_wall->data = expected_data;
	expected_data->next = next_data;
	expected_data->x = x;
	expected_data->y = y;
	expected_data->wall = expected_wall;
	expected_data->flags = (uint8_t)flags;
	expected_data->state = (uint8_t)state;
	expected_data->open_delay = (uint8_t)delay;
	expected_unit->obj_class = unit_class;
	lookup_result = expected_wall;
	sequence = 0;
	sound_calls = 0;
	failure = 0;
	mutate_tile = 0;
	wanted_pos.field_0 = original_center(x);
	wanted_pos.field_4 = original_center(y);
}

static int check_touch(int opening) {
	nox_wall_native_t old_wall = *expected_wall;
	nox_secret_wall_t wanted_data = *expected_data;
	nox_secret_wall_t old_next = *next_data;
	nox_object_t old_unit = *expected_unit;
	int2 old_grid = lookup_grid;
	if (opening) {
		wanted_data.state = 4;
		wanted_data.open_delay = 0;
	}
	if (mutate_tile) old_wall.tile = 255;
	sub_548100(&lookup_grid, expected_unit);
	CHECK(!failure && sequence == (opening ? 1234U : 1U));
	CHECK(sound_calls == (unsigned int)opening);
	CHECK(!memcmp(expected_data, &wanted_data, sizeof(wanted_data)));
	CHECK(!memcmp(expected_wall, &old_wall, sizeof(old_wall)));
	CHECK(!memcmp(expected_unit, &old_unit, sizeof(old_unit)));
	CHECK(!memcmp(next_data, &old_next, sizeof(old_next)));
	CHECK(!memcmp(&lookup_grid, &old_grid, sizeof(old_grid)));
	return 0;
}

int main(void) {
	expected_wall = calloc(1, sizeof(*expected_wall));
	expected_data = calloc(1, sizeof(*expected_data));
	next_data = calloc(1, sizeof(*next_data));
	expected_unit = calloc(1, sizeof(*expected_unit));
	if (!expected_wall || !expected_data || !next_data || !expected_unit) return 2;
	if (sizeof(void*) == 8) {
		CHECK((uintptr_t)expected_wall > UINT32_MAX && (uintptr_t)expected_data > UINT32_MAX &&
			(uintptr_t)next_data > UINT32_MAX && (uintptr_t)expected_unit > UINT32_MAX);
	}
	// This first closed, touch-enabled player case fails with PE32 byte indexing.
	prepare(4, 4, 2, 1, 21, 48, 62);
	CHECK(check_touch(1) == 0);
	// Repeated contact must not restart opening or duplicate its sound.
	sequence = 0;
	sound_calls = 0;
	CHECK(check_touch(0) == 0);
	const unsigned int classes[] = {0, 1, 2, 4, 6, 8, 0x1000000, 0x1000006};
	for (unsigned int state = 0; state < 256; state++) {
		for (unsigned int flags = 0; flags < 256; flags++) {
			for (unsigned int c = 0; c < sizeof(classes) / sizeof(classes[0]); c++) {
				prepare(4, classes[c], flags, state, 255, 48, 62);
				CHECK(check_touch(state == 1 && (flags & 2) && (classes[c] & 6)) == 0);
			}
		}
	}
	for (unsigned int flags = 0; flags < 256; flags++) {
		prepare(flags, 4, 2, 1, 13, 48, 62);
		CHECK(check_touch((flags & 4) != 0) == 0);
	}
	const int32_t coordinates[] = {0, 255, -1, INT32_MIN, INT32_MAX, 0x12345678, -0x12345678};
	for (unsigned int x = 0; x < sizeof(coordinates) / sizeof(coordinates[0]); x++) {
		for (unsigned int y = 0; y < sizeof(coordinates) / sizeof(coordinates[0]); y++) {
			prepare(4, 2, 2, 1, 23, coordinates[x], coordinates[y]);
			mutate_tile = 1;
			CHECK(check_touch(1) == 0);
		}
	}
	prepare(4, 4, 2, 1, 19, 48, 62);
	lookup_result = NULL;
	CHECK(check_touch(0) == 0);
	prepare(4, 4, 2, 1, 19, 48, 62);
	expected_wall->data = NULL;
	CHECK(check_touch(0) == 0);
	// Non-secret or absent walls must not dereference the unit argument.
	prepare(0, 4, 2, 1, 19, 48, 62);
	sub_548100(&lookup_grid, NULL);
	CHECK(!failure && sequence == 1 && !sound_calls && expected_data->state == 1);
	sequence = 0;
	lookup_result = NULL;
	sub_548100(&lookup_grid, NULL);
	CHECK(!failure && sequence == 1 && !sound_calls && expected_data->state == 1);
	free(expected_unit);
	free(next_data);
	free(expected_data);
	free(expected_wall);
	return 0;
}
