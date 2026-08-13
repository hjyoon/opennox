#include "../GAME3_3.h"

#ifdef NOX_UNIT_ITEM_EQUIVALENT_4E7DE0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields and ABI consumed by 004E7DE0.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "object type width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_subclass) == 4, "object subclass width");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_subclass) == (sizeof(void*) == 4 ? 12 : 16), "object subclass offset");
_Static_assert(offsetof(nox_object_t, init_data) == (sizeof(void*) == 4 ? 692 : 760), "init-data offset");
_Static_assert(offsetof(nox_object_t, use_data) == (sizeof(void*) == 4 ? 736 : 848), "use-data offset");
_Static_assert(offsetof(nox_modifier_attrs_t, modifiers) == 0, "modifier-array offset");
_Static_assert(sizeof(((nox_modifier_attrs_t*)0)->modifiers) == 4 * sizeof(void*), "modifier-array width");

static int32_t (*const nox_unit_item_equivalent_signature_4e7de0)(
	const nox_object_t*, const nox_object_t*
) = sub_4E7DE0;
#endif
