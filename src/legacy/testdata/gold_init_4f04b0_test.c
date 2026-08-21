// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../gold_init_4f04b0.h"

#include <limits.h>
#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

typedef struct gold_init_data {
	uint32_t amount;
} gold_init_data;

struct nox_object_t {
	uint32_t guard_before;
	float experience;
	gold_init_data* init_data;
	uint32_t guard_after;
};

typedef struct gold_test_player {
	nox_object_t* unit;
} gold_test_player;

typedef int32_t (*gold_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "GoldInit result must remain an exact dword");
_Static_assert(sizeof(uint32_t) == 4, "GoldInit Amount must remain an exact dword");
_Static_assert(sizeof(float) == 4, "player experience must remain binary32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(gold_init_data) == 4, "GoldInit data size");
_Static_assert(offsetof(gold_init_data, amount) == 0, "GoldInit Amount offset");
_Static_assert(
	_Generic(&nox_xxx_unitInitGold_4F04B0, gold_init_fn: 1, default: 0),
	"GoldInit must use one native object pointer and return an exact int32_t");

static gold_test_player players[3];
static size_t player_count;
static int32_t random_results[2];
static int32_t random_minimums[2];
static int32_t random_maximums[2];
static const char* random_paths[2];
static int32_t random_lines[2];
static size_t random_calls;

static float spill_binary32(double value) {
	volatile float result = (float)value;
	return result;
}

static int32_t dword_bits_to_int32(uint32_t bits) {
	int32_t result;
	memcpy(&result, &bits, sizeof(result));
	return result;
}

static int32_t trunc_signed_qword_low(double value) {
	int64_t whole;
	if (isnan(value) || value >= 0x1p63 || value < -0x1p63)
		return INT32_C(0);
	whole = (int64_t)value;
	return dword_bits_to_int32((uint32_t)(uint64_t)whole);
}

static int32_t logic_random_int(int32_t minimum, int32_t maximum, const char* path, int32_t line) {
	size_t const index = random_calls++;
	if (index >= 2)
		return INT32_C(0);
	random_minimums[index] = minimum;
	random_maximums[index] = maximum;
	random_paths[index] = path;
	random_lines[index] = line;
	return random_results[index];
}

int32_t nox_xxx_unitInitGold_4F04B0(nox_object_t* unit) {
	static char const source[] = "C:\\NoxPost\\src\\Server\\Object\\init\\Init.c";
	gold_init_data* const init_data = unit->init_data;
	float sum = 0.0f;
	int32_t count = 0;
	float average;
	int32_t upper;
	int32_t lower;
	int32_t scaled_random;
	int32_t negative;
	int32_t base_random;
	uint32_t amount;
	size_t index;

	if (init_data->amount != UINT32_C(0))
		return dword_bits_to_int32((uint32_t)(uintptr_t)unit);

	for (index = 0; index < player_count; ++index) {
		nox_object_t* const player_unit = players[index].unit;
		if (player_unit != NULL)
			sum = spill_binary32((double)sum + (double)player_unit->experience);
		count += INT32_C(1);
	}
	average = spill_binary32((double)sum / (double)count);
	upper = trunc_signed_qword_low((double)average * 0x1.47ae147ae147bp-6);
	lower = trunc_signed_qword_low((double)average * 0x1.47ae147ae147bp-7);
	scaled_random = logic_random_int(lower, upper, source, INT32_C(1017));
	negative = trunc_signed_qword_low((double)average * -0x1.47ae147ae147bp-6);
	base_random = logic_random_int(INT32_C(15), INT32_C(30), source, INT32_C(1018));
	amount = (uint32_t)scaled_random - (uint32_t)negative + (uint32_t)base_random;
	init_data->amount = amount;
	return base_random;
}

static void reset_random(int32_t first, int32_t second) {
	random_results[0] = first;
	random_results[1] = second;
	random_minimums[0] = random_minimums[1] = INT32_C(-1);
	random_maximums[0] = random_maximums[1] = INT32_C(-1);
	random_paths[0] = random_paths[1] = NULL;
	random_lines[0] = random_lines[1] = INT32_C(-1);
	random_calls = 0;
}

int main(void) {
	static char const source[] = "C:\\NoxPost\\src\\Server\\Object\\init\\Init.c";
	struct guarded_data {
		gold_init_data data;
		uint32_t guard;
	} record = {{UINT32_C(0)}, UINT32_C(0xA5A5A5A5)};
	nox_object_t first_unit = {UINT32_C(0x11111111), 1200.0f, NULL, UINT32_C(0x22222222)};
	nox_object_t third_unit = {UINT32_C(0x33333333), 1800.0f, NULL, UINT32_C(0x44444444)};
	nox_object_t gold = {UINT32_C(0x55555555), 0.0f, &record.data, UINT32_C(0x66666666)};
	gold_init_fn const init = nox_xxx_unitInitGold_4F04B0;
	int32_t result;

	players[0].unit = &first_unit;
	players[1].unit = NULL;
	players[2].unit = &third_unit;
	player_count = 3;
	reset_random(INT32_C(17), INT32_C(23));
	result = init(&gold);
	if (result != INT32_C(23) || record.data.amount != UINT32_C(60))
		return __LINE__;
	if (random_calls != 2 || random_minimums[0] != INT32_C(10) || random_maximums[0] != INT32_C(20) ||
		random_minimums[1] != INT32_C(15) || random_maximums[1] != INT32_C(30))
		return __LINE__;
	if (random_lines[0] != INT32_C(1017) || random_lines[1] != INT32_C(1018) ||
		strcmp(random_paths[0], source) != 0 || strcmp(random_paths[1], source) != 0)
		return __LINE__;
	if (record.guard != UINT32_C(0xA5A5A5A5) || gold.guard_before != UINT32_C(0x55555555) ||
		gold.guard_after != UINT32_C(0x66666666))
		return __LINE__;

	record.data.amount = UINT32_C(1);
	reset_random(INT32_C(0), INT32_C(0));
	result = init(&gold);
	if (result != dword_bits_to_int32((uint32_t)(uintptr_t)&gold) || record.data.amount != UINT32_C(1) ||
		random_calls != 0)
		return __LINE__;

	record.data.amount = UINT32_C(0);
	player_count = 0;
	reset_random(INT32_C(0), INT32_C(19));
	result = init(&gold);
	if (result != INT32_C(19) || record.data.amount != UINT32_C(19) || random_calls != 2 ||
		random_minimums[0] != INT32_C(0) || random_maximums[0] != INT32_C(0))
		return __LINE__;
	return 0;
}
