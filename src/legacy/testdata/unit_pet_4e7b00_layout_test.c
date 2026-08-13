#include "../GAME3_3.h"

#ifdef NOX_UNIT_PET_4E7B00_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the fields and ABIs used by this unit.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, obj_subclass) == (sizeof(void*) == 4 ? 12 : 16), "object subclass offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872), "object update data offset");
_Static_assert(offsetof(nox_playerInfo, playerUnit) == 2056, "player unit offset");
_Static_assert(offsetof(nox_playerInfo, playerInd) == (sizeof(void*) == 4 ? 2064 : 2068), "player index offset");

static void (*const nox_unit_become_pet_signature_4e7b00)(nox_object_t*, nox_object_t*) =
	nox_xxx_unitBecomePet_4E7B00;
static void (*const nox_unit_become_enemy_signature_4e7b60)(nox_object_t*, nox_object_t*) =
	nox_xxx_monsterRemoveMonitors_4E7B60;
#endif
