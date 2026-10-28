package legacy

/*
#include "spell_teleport_target_native_530a30.h"
typedef int (*nox_teleport_target_duration_fn)(void*);
_Static_assert(_Generic(&sub_530A30_spell_execdur, nox_teleport_target_duration_fn: 1, default: 0), "native TeleportToTarget create ABI");
_Static_assert(_Generic(&nox_xxx_castTTT_530B70, nox_teleport_target_duration_fn: 1, default: 0), "native TeleportToTarget update ABI");
*/
import "C"
