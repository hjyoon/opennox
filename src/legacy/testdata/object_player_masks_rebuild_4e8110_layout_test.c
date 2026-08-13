#include "../GAME3_3.h"

#ifdef NOX_OBJECT_PLAYER_MASKS_REBUILD_4E8110_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the fields and ABI consumed by 004E8110.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, field_35) == (sizeof(void*) == 4 ? 140 : 144), "object field 35 offset");
_Static_assert(offsetof(nox_object_t, field_36) == (sizeof(void*) == 4 ? 144 : 148), "object field 36 offset");
_Static_assert(offsetof(nox_playerInfo, playerUnit) == 2056, "player unit offset");

static nox_object_t* (*const nox_object_player_masks_rebuild_signature_4e8110)(int32_t) = sub_4E8110;
#endif
