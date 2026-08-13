#include "../GAME3_3.h"

#ifdef NOX_OBJECT_IS_PLAYER_4E7BC0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the field and ABI consumed by 004E7BC0.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");

static int (*const nox_object_is_player_signature_4e7bc0)(const nox_object_t*) = sub_4E7BC0;
#endif
