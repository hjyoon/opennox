package legacy

/*
#include "GAME3_3.h"

typedef int (*player_damage_entry)(nox_object_t*, nox_object_t*, nox_object_t*, int, int);
_Static_assert(_Generic(&nox_server_handler_PlayerDamage_4E17B0, player_damage_entry: 1, default: 0), "player damage ABI");

typedef void (*player_damage_carry_entry)(nox_object_t*, int32_t*, float);
_Static_assert(_Generic(&nox_xxx_playerDecrementHPMana_4E20F0, player_damage_carry_entry: 1, default: 0), "player damage carry ABI");

static void player_damage_carry_call(nox_object_t* target, int32_t* output, float amount) {
	player_damage_carry_entry entry = nox_xxx_playerDecrementHPMana_4E20F0;
	entry(target, output, amount);
}

static int player_damage_legacy32_call(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int damage, int damage_type) {
#if UINTPTR_MAX == UINT32_MAX
	return nox_server_playerDamagePE32_4E17B0((int)(intptr_t)target, (int)(intptr_t)source,
		(int)(intptr_t)weapon, damage, damage_type);
#else
	return 0;
#endif
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Retain the same callback identity that thing.bin binds to PlayerDamage.
func playerDamageMeleeCallbackNative4E17B0() unsafe.Pointer {
	return C.nox_server_handler_PlayerDamage_4E17B0
}

func playerDamageLegacy32Call4E17B0(target, source, weapon *nox_object_t, damage, typ int32) int32 {
	return int32(C.player_damage_legacy32_call(target, source, weapon, C.int(damage), C.int(typ)))
}

func playerDamageCarryPublicCall4E20F0(target *server.Object, output *int32, amount float32) {
	C.player_damage_carry_call(asObjectC(target), (*C.int32_t)(output), C.float(amount))
}
