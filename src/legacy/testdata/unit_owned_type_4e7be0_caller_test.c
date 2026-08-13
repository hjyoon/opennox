#include "../GAME3_3.h"

#include <assert.h>

static uint32_t crown_cache_words[3];
static uint32_t gameball_cache_words[3];
static unsigned int memory_calls;
static unsigned int name_calls;
static uint32_t lookup_result;
static nox_object_t* lookup_owner;
static nox_object_t* lookup_first;
static const char* expected_name;

uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t offset) {
	assert(base == UINT32_C(0x5D4594));
	++memory_calls;
	if (offset == UINT32_C(1567716)) {
		return &crown_cache_words[1];
	}
	assert(offset == UINT32_C(1567720));
	return &gameball_cache_words[1];
}

int nox_xxx_getNameId_4E3AA0(char* name) {
	++name_calls;
	assert(strcmp(name, expected_name) == 0);
	if (lookup_owner) {
		lookup_owner->field_129 = lookup_first;
	}
	return (int)lookup_result;
}

#include "../unit_owned_type_4e7be0.c"

static int (*const crown_signature_4e7be0)(const nox_object_t*) = nox_xxx_unitIsCrown_4E7BE0;
static int (*const gameball_signature_4e7c30)(const nox_object_t*) = nox_xxx_unitIsGameball_4E7C30;

static void reset_counters(void) {
	memory_calls = 0;
	name_calls = 0;
	lookup_owner = NULL;
	lookup_first = NULL;
}

static void assert_cache_neighbors(void) {
	assert(crown_cache_words[0] == UINT32_C(0x11223344));
	assert(crown_cache_words[2] == UINT32_C(0x55667788));
	assert(gameball_cache_words[0] == UINT32_C(0x99aabbcc));
	assert(gameball_cache_words[2] == UINT32_C(0xddeeff00));
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t old_item = {0};
	nox_object_t crown_item = {0};
	nox_object_t ball_item = {0};
	nox_object_t tail = {0};

	crown_cache_words[0] = UINT32_C(0x11223344);
	crown_cache_words[1] = 0;
	crown_cache_words[2] = UINT32_C(0x55667788);
	gameball_cache_words[0] = UINT32_C(0x99aabbcc);
	gameball_cache_words[1] = 0;
	gameball_cache_words[2] = UINT32_C(0xddeeff00);

	old_item.typ_ind = UINT16_C(1);
	crown_item.typ_ind = UINT16_C(0x2468);
	ball_item.typ_ind = UINT16_C(0x1357);
	tail.typ_ind = UINT16_C(0xabcd);
	owner.field_129 = &old_item;

	reset_counters();
	expected_name = "Crown";
	lookup_result = UINT32_C(0x2468);
	lookup_owner = &owner;
	lookup_first = &crown_item;
	assert(crown_signature_4e7be0(&owner) == 1);
	assert(memory_calls == 1);
	assert(name_calls == 1);
	assert(crown_cache_words[1] == UINT32_C(0x2468));
	assert(owner.field_129 == &crown_item);
	assert_cache_neighbors();

	reset_counters();
	owner.field_129 = &crown_item;
	expected_name = NULL;
	assert(crown_signature_4e7be0(&owner) == 1);
	assert(memory_calls == 1);
	assert(name_calls == 0);
	assert_cache_neighbors();

	reset_counters();
	owner.field_129 = &old_item;
	old_item.field_128 = &ball_item;
	ball_item.field_128 = &tail;
	expected_name = "GameBall";
	lookup_result = UINT32_C(0x1357);
	assert(gameball_signature_4e7c30(&owner) == 1);
	assert(memory_calls == 1);
	assert(name_calls == 1);
	assert(gameball_cache_words[1] == UINT32_C(0x1357));
	assert(old_item.field_128 == &ball_item);
	assert(ball_item.field_128 == &tail);
	assert_cache_neighbors();

	reset_counters();
	crown_cache_words[1] = UINT32_C(0x00010001);
	owner.field_129 = &old_item;
	old_item.field_128 = NULL;
	assert(crown_signature_4e7be0(&owner) == 0);
	assert(memory_calls == 1);
	assert(name_calls == 0);
	assert_cache_neighbors();

	reset_counters();
	gameball_cache_words[1] = 0;
	ball_item.typ_ind = 0;
	ball_item.field_128 = NULL;
	owner.field_129 = &ball_item;
	expected_name = "GameBall";
	lookup_result = 0;
	assert(gameball_signature_4e7c30(&owner) == 1);
	assert(gameball_signature_4e7c30(&owner) == 1);
	assert(memory_calls == 2);
	assert(name_calls == 2);
	assert(gameball_cache_words[1] == 0);
	assert_cache_neighbors();

	reset_counters();
	crown_cache_words[1] = UINT32_C(9);
	owner.field_129 = NULL;
	assert(crown_signature_4e7be0(&owner) == 0);
	assert(memory_calls == 1);
	assert(name_calls == 0);
	assert_cache_neighbors();

	return 0;
}
