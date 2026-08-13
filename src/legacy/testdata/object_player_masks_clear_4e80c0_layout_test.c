#include "../GAME3_3.h"

#ifdef NOX_OBJECT_PLAYER_MASKS_CLEAR_4E80C0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the object fields consumed by 004E80C0.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, field_35) == (sizeof(void*) == 4 ? 140 : 144), "object field 35 offset");
_Static_assert(offsetof(nox_object_t, field_36) == (sizeof(void*) == 4 ? 144 : 148), "object field 36 offset");
#endif
