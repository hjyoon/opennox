#include "../GAME3_3.h"

#ifdef NOX_UNIT_INVENTORY_CONTAINS_EQUIVALENT_4E7EC0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields and ABI consumed by 004E7EC0.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528), "inventory-next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544), "first-inventory offset");

static int32_t (*const nox_unit_inventory_contains_equivalent_signature_4e7ec0)(
	const nox_object_t*, const nox_object_t*
) = sub_4E7EC0;
#endif
