#include "../defs.h"

#ifdef NOX_UNIT_COUNT_INVENTORY_CLASS_SUBCLASS_4E7DA0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields consumed by 004E7DA0.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_subclass) == 4, "object subclass width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_subclass) == (sizeof(void*) == 4 ? 12 : 16), "object subclass offset");
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528), "inventory-next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544), "first-inventory offset");
#endif
