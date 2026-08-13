#include "../GAME3_3.h"

#ifdef NOX_MONSTER_MARK_UPDATE_4E8020_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the fields and ABI consumed by 004E8020.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, field_35) == (sizeof(void*) == 4 ? 140 : 144), "object field 35 offset");
_Static_assert(offsetof(nox_object_t, field_36) == (sizeof(void*) == 4 ? 144 : 148), "object field 36 offset");
_Static_assert(offsetof(nox_playerInfo, playerUnit) == 2056, "player unit offset");
_Static_assert(offsetof(nox_playerInfo, playerInd) == (sizeof(void*) == 4 ? 2064 : 2068), "player index offset");

static void (*const nox_monster_mark_update_signature_4e8020)(nox_object_t*) =
	nox_xxx_monsterMarkUpdate_4E8020;
#endif
