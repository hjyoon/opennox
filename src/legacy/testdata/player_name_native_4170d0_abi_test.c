#include <stddef.h>

#include "../GAME1.h"

typedef nox_playerInfo* (*player_by_name_fn)(wchar2_t*);

_Static_assert(_Generic(&nox_xxx_playerByName_4170D0, player_by_name_fn: 1, default: 0),
	"004170D0 must return a native-width player pointer");
_Static_assert(sizeof(((nox_playerInfo*)0)->playerUnit) == sizeof(void*),
	"playerUnit must remain a native-width pointer");
_Static_assert(offsetof(nox_playerInfo, name_final) > offsetof(nox_playerInfo, playerUnit),
	"player name lookup must use named native-layout fields");

int main(void) {
	return 0;
}
