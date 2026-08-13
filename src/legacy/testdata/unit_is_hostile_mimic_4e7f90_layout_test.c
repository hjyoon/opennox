#include "../GAME3_3.h"

#ifdef NOX_UNIT_IS_HOSTILE_MIMIC_4E7F90_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the fields and ABI consumed by 004E7F90.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "object type-index width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type-index offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552), "object owner offset");

static int32_t (*const nox_is_hostile_mimic_signature_4e7f90)(nox_object_t*, nox_object_t*) =
	nox_xxx_unitIsHostileMimic_4E7F90;
#endif
