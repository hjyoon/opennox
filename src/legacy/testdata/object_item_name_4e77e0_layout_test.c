#include "../server__object__objutil.h"

#ifdef NOX_ITEM_NAME_4E77E0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable the exact object and modifier fields used by 004E77E0.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type index offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, init_data) == (sizeof(void*) == 4 ? 692 : 760), "object init data offset");

_Static_assert(offsetof(nox_modifier_attrs_t, modifiers) == 0, "modifier array offset");
_Static_assert(offsetof(nox_modifier_attrs_t, field_16) == (sizeof(void*) == 4 ? 16 : 32), "modifier tail offset");
_Static_assert(sizeof(nox_modifier_attrs_t) == (sizeof(void*) == 4 ? 20 : 40), "modifier attributes size");

_Static_assert(offsetof(obj_412ae0_t, field_2) == (sizeof(void*) == 4 ? 8 : 16), "modifier description offset");
_Static_assert(offsetof(obj_412ae0_t, field_4) == (sizeof(void*) == 4 ? 16 : 32), "modifier identity description offset");
_Static_assert(sizeof(obj_412ae0_t) == (sizeof(void*) == 4 ? 144 : 208), "modifier effect size");

static wchar2_t* (*const nox_item_name_signature_4e77e0)(nox_object_t*) =
	nox_xxx_itemGetName_4E77E0_obj_util;
#endif
