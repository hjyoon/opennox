#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_level_set_4ef410.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef void (*player_level_set_fn)(nox_object_t*, uint8_t);

_Static_assert(CHAR_BIT == 8, "level bytes must remain eight bits");
_Static_assert(sizeof(uint8_t) == 1, "level must remain exact uint8");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EF410, player_level_set_fn: 1, default: 0),
	"player level setter must receive a native object pointer and exact uint8 level");

static nox_object_t* observed_unit;
static uint8_t observed_level;
static unsigned int observed_calls;

void sub_4EF410(nox_object_t* unit, uint8_t level) {
	observed_unit = unit;
	observed_level = level;
	++observed_calls;
}

static void check_call(
		player_level_set_fn set_level,
		nox_object_t* unit,
		uint8_t level) {
	set_level(unit, level);
	assert(observed_unit == unit);
	assert(observed_level == level);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_level_set_fn const set_level = sub_4EF410;

	check_call(set_level, &unit, UINT8_C(0));
	check_call(set_level, &unit, UINT8_C(10));
	check_call(set_level, &unit, UINT8_C(127));
	check_call(set_level, &unit, UINT8_C(128));
	check_call(set_level, NULL, UINT8_MAX);
	assert(observed_calls == 5);
	assert(observed_unit == NULL);
	return 0;
}
