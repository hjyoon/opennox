#include "../GAME3_3.h"

#ifdef NOX_UNIT_COUNT_SLAVES_4E7CF0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the fields and ABI consumed by 004E7CF0.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_subclass) == 4, "object subclass width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_subclass) == (sizeof(void*) == 4 ? 12 : 16), "object subclass offset");
_Static_assert(offsetof(nox_object_t, field_128) == (sizeof(void*) == 4 ? 512 : 560), "owned-next offset");
_Static_assert(offsetof(nox_object_t, field_129) == (sizeof(void*) == 4 ? 516 : 568), "first-owned offset");

static int32_t (*const nox_unit_count_slaves_signature_4e7cf0)(
	const nox_object_t*, uint32_t, uint32_t
) = nox_xxx_unitCountSlaves_4E7CF0;
#endif
