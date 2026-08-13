#include "../GAME3_3.h"

#ifdef NOX_UNIT_COUNT_INVENTORY_TYPE_4E7D30_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields and ABI consumed by 004E7D30.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "object type-index width");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type-index offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528), "inventory-next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544), "first-inventory offset");

static int32_t (*const nox_count_inventory_type_signature_4e7d30)(nox_object_t*, int32_t) =
	nox_xxx_inventoryCountObjects_4E7D30;
#endif
