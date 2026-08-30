// Suppress unrelated Win32-only assertions while parsing the shared header,
// then probe the native-width PlayerUpdate equipped-weapon slot directly.
#define _Static_assert(...)
#include "../defs.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_player_update_data_t, equipped_weapon) == 104,
	"PlayerUpdate equipped-weapon offset");
_Static_assert(sizeof(((nox_player_update_data_t*)0)->equipped_weapon) == sizeof(void*),
	"PlayerUpdate equipped-weapon width");
_Static_assert(offsetof(nox_player_update_data_t, field_59_0) ==
	(sizeof(void*) == 4 ? 236 : 296), "PlayerUpdate later-field stability");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 336), "PlayerUpdate player offset");
_Static_assert(sizeof(nox_player_update_data_t) ==
	(sizeof(void*) == 4 ? 320 : 416), "partial PlayerUpdate size");

int main(void) {
	nox_player_update_data_t update = {0};
	nox_object_t weapon = {0};
	update.equipped_weapon = &weapon;
	return update.equipped_weapon == &weapon ? 0 : 1;
}
