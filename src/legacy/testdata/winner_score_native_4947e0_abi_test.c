#include <stddef.h>
#include <stdint.h>

#include "../GAME2_3.h"

typedef void (*player_winner_score_fn)(nox_playerInfo*);
typedef void (*team_winner_score_fn)(nox_team_t*);

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"winner score objects must remain native-width");
_Static_assert(_Generic(&sub_4947E0, player_winner_score_fn: 1, default: 0),
	"004947E0 must accept native-width player pointers");
_Static_assert(_Generic(&sub_4948B0, team_winner_score_fn: 1, default: 0),
	"004948B0 must accept native-width team pointers");
_Static_assert(offsetof(nox_playerInfo, lessons) == (sizeof(void*) == 4 ? 2136 : 2140),
	"winner scoring must use the native player score layout");
_Static_assert(offsetof(nox_playerInfo, field_2140) == (sizeof(void*) == 4 ? 2140 : 2144),
	"Highlander winner scoring must use the native player score layout");
_Static_assert(offsetof(nox_playerInfo, field_3680) == (sizeof(void*) == 4 ? 3680 : 4976),
	"winner scoring must use the native player status layout");
_Static_assert(offsetof(nox_playerInfo, name_final) == (sizeof(void*) == 4 ? 4704 : 6008),
	"winner announcements must use the native player name layout");
_Static_assert(offsetof(nox_playerInfo, info) + offsetof(nox_playerInfo2, isFemale) ==
		(sizeof(void*) == 4 ? 2252 : 2256),
	"winner announcements must use the native player gender layout");
_Static_assert(offsetof(nox_object_t, field_12) == (sizeof(void*) == 4 ? 48 : 52),
	"server winner scoring must use the native object team layout");
_Static_assert(offsetof(nox_drawable, field_6) == (sizeof(void*) == 4 ? 24 : 36),
	"client winner scoring must use the native drawable team layout");
_Static_assert(offsetof(nox_team_t, field_57) == 57,
	"winner scoring must preserve the fixed team identifier offset");

int main(void) {
	return 0;
}
