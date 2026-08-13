#include "../GAME3_2.h"
#include "../GAME3_3.h"

#ifdef NOX_UNIT_FREEZE_4E79C0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the object fields and ABIs used by this unit.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, field_128) == (sizeof(void*) == 4 ? 512 : 560), "owned next offset");
_Static_assert(offsetof(nox_object_t, field_129) == (sizeof(void*) == 4 ? 516 : 568), "owned first offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872), "object update data offset");

static int (*const nox_player_status_signature_4d8270)(nox_object_t*) =
	nox_xxx_netReportPlrStatus_4D8270;
static uint32_t (*const nox_freeze_gate_set_signature_4e79b0)(uint32_t) =
	sub_4E79B0;
static uint32_t* (*const nox_freeze_gate_ref_signature_4e79b0)(void) =
	nox_xxx_unitFreezeGateRef_4E79B0;
static char (*const nox_unit_freeze_signature_4e79c0)(nox_object_t*, uint32_t) =
	nox_xxx_unitFreeze_4E79C0;
static char (*const nox_unit_unfreeze_signature_4e7a60)(nox_object_t*, uint32_t) =
	nox_xxx_unitUnFreeze_4E7A60;
#endif
