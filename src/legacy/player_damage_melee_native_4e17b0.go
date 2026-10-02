package legacy

/*
#include "defs.h"
int nox_server_handler_PlayerDamage_4E17B0_go(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type);
*/
import "C"

import "unsafe"

// Retain the same callback identity that thing.bin binds to PlayerDamage.
func playerDamageMeleeCallbackNative4E17B0() unsafe.Pointer {
	return C.nox_server_handler_PlayerDamage_4E17B0_go
}
