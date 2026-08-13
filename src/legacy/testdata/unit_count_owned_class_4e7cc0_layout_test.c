#include "../defs.h"

#ifdef NOX_UNIT_COUNT_OWNED_CLASS_4E7CC0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields consumed by 004E7CC0.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, field_128) == (sizeof(void*) == 4 ? 512 : 560), "owned-next offset");
_Static_assert(offsetof(nox_object_t, field_129) == (sizeof(void*) == 4 ? 516 : 568), "first-owned offset");
#endif
