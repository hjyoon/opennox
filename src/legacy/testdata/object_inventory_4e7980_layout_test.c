#include "../GAME3_3.h"

#ifdef NOX_INVENTORY_4E7980_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable the fields and signatures consumed by 004E7980/90.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528), "inventory next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544), "inventory first offset");

static nox_object_t* (*const nox_inventory_first_signature_4e7980)(nox_object_t*) =
	nox_xxx_inventoryGetFirst_4E7980;
static nox_object_t* (*const nox_inventory_next_signature_4e7990)(nox_object_t*) =
	nox_xxx_inventoryGetNext_4E7990;
#endif
