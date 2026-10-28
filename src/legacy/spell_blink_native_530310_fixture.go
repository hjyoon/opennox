package legacy

/*
#include "spell_blink_native_530310.h"
typedef int (*nox_blink_duration_fn)(void*);
typedef void (*nox_blink_wake_fn)(void*, void*, void*);
_Static_assert(_Generic(&nox_xxx_spellBlink2_530310, nox_blink_duration_fn: 1, default: 0), "native Blink create ABI");
_Static_assert(_Generic(&nox_xxx_spellBlink1_530380, nox_blink_duration_fn: 1, default: 0), "native Blink update ABI");
_Static_assert(_Generic(&nox_xxx_spellTeleportCreateWake_530560, nox_blink_wake_fn: 1, default: 0), "native TeleportWake ABI");
*/
import "C"
