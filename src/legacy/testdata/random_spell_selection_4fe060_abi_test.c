#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../random_spell_selection_4fe060.h"

typedef int32_t (*random_spell_selection_fn)(uint32_t, uint32_t);
typedef int32_t (*random_spell_excluded_fn)(int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "spell masks must remain unsigned dwords");
_Static_assert(sizeof(int32_t) == 4, "spell IDs and results must remain signed dwords");
_Static_assert(
	_Generic(&nox_xxx_unused_4FE060, random_spell_selection_fn: 1, default: 0),
	"004FE060 must preserve two unsigned dword masks and a signed dword result");
_Static_assert(
	_Generic(&sub_4FE100, random_spell_excluded_fn: 1, default: 0),
	"004FE100 must preserve its signed dword ID and result");

static uint32_t observed_first_mask;
static uint32_t observed_second_mask;
static int32_t next_result;

int32_t nox_xxx_unused_4FE060(uint32_t first_mask, uint32_t second_mask) {
	observed_first_mask = first_mask;
	observed_second_mask = second_mask;
	return next_result;
}

int32_t sub_4FE100(int32_t spell_id) {
	switch (spell_id) {
	case 1: case 2: case 6: case 13: case 15: case 18:
	case 19: case 20: case 30: case 32: case 33: case 34:
	case 38: case 51: case 57: case 68: case 69: case 70:
	case 73: case 129: case 133:
		return 1;
	default:
		return 0;
	}
}

int main(void) {
	random_spell_selection_fn const select_spell = nox_xxx_unused_4FE060;
	random_spell_excluded_fn const excluded = sub_4FE100;
	static const int32_t excluded_ids[] = {
		1, 2, 6, 13, 15, 18, 19, 20, 30, 32, 33, 34,
		38, 51, 57, 68, 69, 70, 73, 129, 133,
	};

	next_result = INT32_MIN;
	assert(select_spell(UINT32_MAX, UINT32_C(0x80000000)) == INT32_MIN);
	assert(observed_first_mask == UINT32_MAX);
	assert(observed_second_mask == UINT32_C(0x80000000));

	next_result = INT32_MAX;
	assert(select_spell(0, UINT32_MAX) == INT32_MAX);
	assert(observed_first_mask == 0);
	assert(observed_second_mask == UINT32_MAX);

	for (size_t i = 0; i < sizeof(excluded_ids) / sizeof(excluded_ids[0]); ++i) {
		assert(excluded(excluded_ids[i]) == 1);
	}
	assert(excluded(INT32_MIN) == 0);
	assert(excluded(-1) == 0);
	assert(excluded(0) == 0);
	assert(excluded(134) == 0);
	assert(excluded(INT32_MAX) == 0);
	return 0;
}
