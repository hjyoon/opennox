#include <limits.h>
#include <stdint.h>

#include "../GAME1_2.h"
#include "../GAME3.h"
#include "../client__shell__noxworld.h"

typedef int (*game_list_event_fn)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*game_list_event_unsigned_fn)(nox_window*, unsigned int, uintptr_t, uintptr_t);
typedef int (*game_list_first_visible_fn)(const nox_scrollListBox_data*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"game-list GUI pointers must remain native-width");
_Static_assert(_Generic(&sub_438EF0, game_list_event_fn: 1, default: 0),
	"00438EF0 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_439050, game_list_event_unsigned_fn: 1, default: 0),
	"00439050 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&nox_xxx_windowMultiplayerSub_439E70,
	game_list_event_unsigned_fn: 1, default: 0),
	"00439E70 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_4A4800, game_list_first_visible_fn: 1, default: 0),
	"004A4800 must consume the native scroll-list data layout");

int main(void) {
	return 0;
}
