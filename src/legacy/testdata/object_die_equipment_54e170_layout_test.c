#include "../server__object__die__die.h"

#ifdef NOX_EQUIPMENT_DIE_54E170_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable every field consumed by 0054E170 and 0054E370.
#undef _Static_assert
_Static_assert(sizeof(int) == 4, "32-bit legacy integer width");
_Static_assert(sizeof(wchar2_t) == 2, "wide character width");
_Static_assert(sizeof(float2) == 8, "position pair size");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type index offset");
_Static_assert(offsetof(nox_object_t, material) == (sizeof(void*) == 4 ? 24 : 28), "object material offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, inv_holder) == (sizeof(void*) == 4 ? 492 : 520), "object holder offset");
_Static_assert(offsetof(obj_412ae0_t, field_2) == (sizeof(void*) == 4 ? 8 : 16), "definition description offset");

static void (*const nox_armor_die_signature_54e170)(nox_object_t*) = nox_xxx_dieArmor_54E170_obj_die;
static void (*const nox_weapon_die_signature_54e370)(nox_object_t*) = nox_xxx_dieWeapon_54E370_obj_die;
#endif
