package legacy

/*
#include "spell_tag_native_530160.h"
typedef int (*nox_tag_create_fn)(void*);
typedef unsigned int (*nox_tag_update_fn)(void*);
typedef void (*nox_tag_destroy_fn)(void*);
_Static_assert(_Generic(&nox_xxx_spellTagCreature_530160, nox_tag_create_fn: 1, default: 0), "native Tag create ABI");
_Static_assert(_Generic(&sub_530250, nox_tag_update_fn: 1, default: 0), "native Tag update ABI");
_Static_assert(_Generic(&sub_530270, nox_tag_destroy_fn: 1, default: 0), "native Tag destroy ABI");
*/
import "C"
