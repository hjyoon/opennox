#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_sync_level_4ef140.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*player_sync_level_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "level bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "result must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EF140, player_sync_level_fn: 1, default: 0),
	"player level synchronization must use a native object pointer and exact int32 result");

static nox_object_t* observed_unit;
static int32_t next_result;
static unsigned int observed_calls;

int32_t sub_4EF140(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
	return next_result;
}

static void check_call(
		player_sync_level_fn sync_level,
		nox_object_t* unit,
		int32_t result) {
	next_result = result;
	assert(sync_level(unit) == result);
	assert(observed_unit == unit);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_sync_level_fn const sync_level = sub_4EF140;

	check_call(sync_level, &unit, INT32_C(0));
	check_call(sync_level, &unit, INT32_MAX);
	check_call(sync_level, &unit, INT32_MIN);
	check_call(sync_level, NULL, INT32_C(-1));
	assert(observed_calls == 4);
	assert(observed_unit == NULL);
	return 0;
}
