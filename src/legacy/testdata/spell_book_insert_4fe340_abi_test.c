#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../spell_book_insert_4fe340.h"

typedef int32_t (*spell_book_insert_fn)(nox_object_t*, int32_t*, int32_t, int32_t, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell words, scalar arguments, and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_spellByBookInsert_4FE340, spell_book_insert_fn: 1, default: 0),
	"004FE340 must preserve two native pointers and three signed dword arguments/result");

struct nox_object_t {
	uint32_t marker;
};

static nox_object_t* observed_unit;
static int32_t* observed_sequence;
static int32_t observed_args[3];

int32_t nox_xxx_spellByBookInsert_4FE340(
	nox_object_t* unit,
	int32_t* sequence,
	int32_t count,
	int32_t delay,
	int32_t target_mode) {
	observed_unit = unit;
	observed_sequence = sequence;
	observed_args[0] = count;
	observed_args[1] = delay;
	observed_args[2] = target_mode;
	return INT32_MIN;
}

int main(void) {
	nox_object_t unit = {0};
	int32_t sequence[5] = {INT32_MIN, -1, 0, 1, INT32_MAX};
	spell_book_insert_fn const insert = nox_xxx_spellByBookInsert_4FE340;

	assert(insert(&unit, sequence, INT32_MIN, -1234567, INT32_MAX) == INT32_MIN);
	assert(observed_unit == &unit);
	assert(observed_sequence == sequence);
	assert(observed_args[0] == INT32_MIN);
	assert(observed_args[1] == -1234567);
	assert(observed_args[2] == INT32_MAX);
	return 0;
}
