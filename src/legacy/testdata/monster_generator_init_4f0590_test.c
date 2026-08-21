// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../monster_generator_init_4f0590.h"

#include <limits.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

typedef struct monster_generator_update_data {
	uint8_t quest_spawn_rate[3];
	uint8_t active_count;
	uint8_t max_active;
	uint8_t guard[3];
} monster_generator_update_data;

struct nox_object_t {
	uint32_t subclass;
	uint16_t direction1;
	uint16_t direction2;
	monster_generator_update_data* update_data;
	uint32_t guard;
};

typedef int32_t (*monster_generator_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "MonsterGeneratorInit result must remain an exact dword");
_Static_assert(sizeof(uint32_t) == 4, "subclass must remain an exact dword");
_Static_assert(sizeof(uint16_t) == 2, "direction must remain an exact word");
_Static_assert(sizeof(float) == 4, "balance values must remain binary32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(offsetof(monster_generator_update_data, quest_spawn_rate) == 0, "selector offset");
_Static_assert(offsetof(monster_generator_update_data, active_count) == 3, "active-count offset");
_Static_assert(offsetof(monster_generator_update_data, max_active) == 4, "max-active offset");
_Static_assert(
	_Generic(&nox_xxx_unitInitGenerator_4F0590, monster_generator_init_fn: 1, default: 0),
	"MonsterGeneratorInit must use one native object pointer and return an exact int32_t");

static uint32_t current_group;
static float balance_value;
static char const* balance_key;
static uint32_t balance_calls;
static uint32_t direction_calls;
static uint32_t direction_index;

static int32_t dword_bits_to_int32(uint32_t bits) {
	int32_t result;
	memcpy(&result, &bits, sizeof(result));
	return result;
}

static int32_t trunc_signed_qword_low(float value) {
	double const widened = (double)value;
	int64_t whole;
	if (isnan(widened) || widened >= 0x1p63 || widened < -0x1p63)
		return INT32_C(0);
	whole = (int64_t)widened;
	return dword_bits_to_int32((uint32_t)(uint64_t)whole);
}

static float load_balance(char const* key) {
	balance_key = key;
	balance_calls += UINT32_C(1);
	return balance_value;
}

static uint32_t direction_index_to_angle(uint32_t index) {
	static uint32_t const table[9] = {
		UINT32_C(160), UINT32_C(192), UINT32_C(224),
		UINT32_C(128), UINT32_C(0), UINT32_C(0),
		UINT32_C(96), UINT32_C(64), UINT32_C(32),
	};
	direction_calls += UINT32_C(1);
	direction_index = index;
	return table[index];
}

int32_t nox_xxx_unitInitGenerator_4F0590(nox_object_t* unit) {
	static char const* const keys[4] = {
		"GeneratorMaxActiveCreaturesHigh",
		"GeneratorMaxActiveCreaturesNormal",
		"GeneratorMaxActiveCreaturesLow",
		"GeneratorMaxActiveCreaturesSingular",
	};
	monster_generator_update_data* const update = unit->update_data;
	uint8_t const selector = update->quest_spawn_rate[current_group];
	uint32_t result;
	uint32_t index = UINT32_C(0);
	int matched = 1;

	if (selector <= UINT8_C(3))
		update->max_active = (uint8_t)trunc_signed_qword_low(load_balance(keys[selector]));

	result = unit->subclass;
	if ((result & UINT32_C(1)) != 0)
		index = UINT32_C(0);
	else if ((result & UINT32_C(2)) != 0)
		index = UINT32_C(2);
	else if ((result & UINT32_C(4)) != 0)
		index = UINT32_C(8);
	else if ((result & UINT32_C(8)) != 0)
		index = UINT32_C(6);
	else
		matched = 0;
	if (matched) {
		result = direction_index_to_angle(index);
		unit->direction1 = (uint16_t)result;
	}
	unit->direction2 = unit->direction1;
	return dword_bits_to_int32(result);
}

static void reset_calls(void) {
	balance_key = NULL;
	balance_calls = UINT32_C(0);
	direction_calls = UINT32_C(0);
	direction_index = UINT32_MAX;
}

int main(void) {
	static char const* const keys[4] = {
		"GeneratorMaxActiveCreaturesHigh",
		"GeneratorMaxActiveCreaturesNormal",
		"GeneratorMaxActiveCreaturesLow",
		"GeneratorMaxActiveCreaturesSingular",
	};
	struct guarded_update {
		monster_generator_update_data data;
		uint32_t guard;
	} record = {{{UINT8_C(0), UINT8_C(1), UINT8_C(2)}, UINT8_C(0x44), UINT8_C(0xAA), {0, 0, 0}}, UINT32_C(0xA5A5A5A5)};
	nox_object_t unit = {UINT32_C(4), UINT16_C(0xAAAA), UINT16_C(0xBBBB), &record.data, UINT32_C(0x5A5A5A5A)};
	monster_generator_init_fn const init = nox_xxx_unitInitGenerator_4F0590;
	int32_t result;
	uint32_t selector;

	for (selector = 0; selector < UINT32_C(4); ++selector) {
		current_group = UINT32_C(1);
		record.data.quest_spawn_rate[1] = (uint8_t)selector;
		record.data.max_active = UINT8_C(0xAA);
		balance_value = selector == UINT32_C(2) ? -1.75f : (float)(256 + selector);
		unit.subclass = UINT32_C(0);
		unit.direction1 = UINT16_C(0x1234);
		unit.direction2 = UINT16_C(0xBBBB);
		reset_calls();
		result = init(&unit);
		if (result != INT32_C(0) || balance_calls != UINT32_C(1) || strcmp(balance_key, keys[selector]) != 0)
			return __LINE__;
		if (record.data.max_active != (selector == UINT32_C(2) ? UINT8_MAX : (uint8_t)selector))
			return __LINE__;
		if (direction_calls != UINT32_C(0) || unit.direction1 != UINT16_C(0x1234) || unit.direction2 != UINT16_C(0x1234))
			return __LINE__;
	}

	record.data.quest_spawn_rate[1] = UINT8_C(4);
	record.data.max_active = UINT8_C(0xAA);
	unit.subclass = UINT32_C(3);
	reset_calls();
	result = init(&unit);
	if (result != INT32_C(160) || record.data.max_active != UINT8_C(0xAA) || balance_calls != UINT32_C(0) ||
		direction_calls != UINT32_C(1) || direction_index != UINT32_C(0) ||
		unit.direction1 != UINT16_C(160) || unit.direction2 != UINT16_C(160))
		return __LINE__;

	unit.subclass = UINT32_C(6);
	reset_calls();
	result = init(&unit);
	if (result != INT32_C(224) || direction_index != UINT32_C(2))
		return __LINE__;

	unit.subclass = UINT32_C(12);
	reset_calls();
	result = init(&unit);
	if (result != INT32_C(32) || direction_index != UINT32_C(8))
		return __LINE__;

	unit.subclass = UINT32_C(8);
	reset_calls();
	result = init(&unit);
	if (result != INT32_C(96) || direction_index != UINT32_C(6))
		return __LINE__;

	unit.subclass = UINT32_C(0x89ABCD00);
	unit.direction1 = UINT16_C(0x4321);
	unit.direction2 = UINT16_C(0xBBBB);
	reset_calls();
	result = init(&unit);
	if (result != dword_bits_to_int32(UINT32_C(0x89ABCD00)) || direction_calls != UINT32_C(0) ||
		unit.direction1 != UINT16_C(0x4321) || unit.direction2 != UINT16_C(0x4321))
		return __LINE__;
	if (record.data.active_count != UINT8_C(0x44) || record.guard != UINT32_C(0xA5A5A5A5) || unit.guard != UINT32_C(0x5A5A5A5A))
		return __LINE__;
	return 0;
}
