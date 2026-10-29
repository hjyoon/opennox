package legacy

/*
#include "spell_turn_undead_native_531310.h"
typedef int (*nox_turn_undead_create_fn)(void*);
typedef void (*nox_turn_undead_destroy_fn)(void*);
_Static_assert(_Generic(&nox_xxx_spellTurnUndeadCreate_531310, nox_turn_undead_create_fn:1, default:0), "native TurnUndead create ABI");
_Static_assert(_Generic(&nox_xxx_spellTurnUndeadDelete_531420, nox_turn_undead_destroy_fn:1, default:0), "native TurnUndead destroy ABI");
*/
import "C"
