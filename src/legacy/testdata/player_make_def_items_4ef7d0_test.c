// Keep this fixture independent from the Win32-only aggregate legacy headers
// so the retained ABI can be compiled by every supported target frontend.
#include "../player_make_def_items_4ef7d0.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef uint8_t (*player_make_def_items_fn)(nox_object_t*, int32_t, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint8_t) == 1, "result must remain exact uint8");
_Static_assert(sizeof(int32_t) == 4, "control arguments must remain exact int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_playerMakeDefItems_4EF7D0, player_make_def_items_fn: 1, default: 0),
	"default-player-item creation must use a native pointer, int32 controls, and uint8 result");

static nox_object_t* observed_player;
static int32_t observed_restore_stats;
static int32_t observed_keep_items;
static uint8_t next_result;
static unsigned int observed_calls;

uint8_t nox_xxx_playerMakeDefItems_4EF7D0(
		nox_object_t* player,
		int32_t restore_stats,
		int32_t keep_items) {
	observed_player = player;
	observed_restore_stats = restore_stats;
	observed_keep_items = keep_items;
	++observed_calls;
	return next_result;
}

static int check_call(
		player_make_def_items_fn make_items,
		nox_object_t* player,
		int32_t restore_stats,
		int32_t keep_items,
		uint8_t result) {
	next_result = result;
	if (make_items(player, restore_stats, keep_items) != result)
		return __LINE__;
	if (observed_player != player)
		return __LINE__;
	if (observed_restore_stats != restore_stats)
		return __LINE__;
	if (observed_keep_items != keep_items)
		return __LINE__;
	return 0;
}

int main(void) {
	unsigned char player_storage = 0;
	nox_object_t* const player = (nox_object_t*)(void*)&player_storage;
	player_make_def_items_fn const make_items = nox_xxx_playerMakeDefItems_4EF7D0;
	int line;

	line = check_call(make_items, player, INT32_MIN, INT32_MAX, UINT8_C(0));
	if (line != 0)
		return line;
	line = check_call(make_items, player, INT32_MAX, INT32_MIN, UINT8_C(127));
	if (line != 0)
		return line;
	line = check_call(make_items, NULL, INT32_C(1), INT32_C(0), UINT8_C(128));
	if (line != 0)
		return line;
	line = check_call(make_items, NULL, INT32_C(0), INT32_C(1), UINT8_MAX);
	if (line != 0)
		return line;
	if (observed_calls != 4)
		return __LINE__;
	return 0;
}
