#include "../GAME3_3.h"

#ifdef NOX_UNIT_POST_CREATE_NOTIFY_4E7F10_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the fields consumed by 004E7F10.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, field_35) == (sizeof(void*) == 4 ? 140 : 144), "object field 35 offset");
_Static_assert(offsetof(nox_object_t, field_36) == (sizeof(void*) == 4 ? 144 : 148), "object field 36 offset");
_Static_assert(offsetof(nox_playerInfo, playerUnit) == 2056, "player unit offset");
_Static_assert(offsetof(nox_playerInfo, playerInd) == (sizeof(void*) == 4 ? 2064 : 2068), "player index offset");
#endif
