#include "../GAME3_3.h"

#ifdef NOX_UNIT_OWNED_TYPE_4E7BE0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields and ABI consumed by this pair.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "object type-index width");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type-index offset");
_Static_assert(offsetof(nox_object_t, field_128) == (sizeof(void*) == 4 ? 512 : 560), "owned-next offset");
_Static_assert(offsetof(nox_object_t, field_129) == (sizeof(void*) == 4 ? 516 : 568), "first-owned offset");

static int (*const nox_crown_signature_4e7be0)(const nox_object_t*) = nox_xxx_unitIsCrown_4E7BE0;
static int (*const nox_gameball_signature_4e7c30)(const nox_object_t*) = nox_xxx_unitIsGameball_4E7C30;
#endif
