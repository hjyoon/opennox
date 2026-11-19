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

typedef double (*player_damage_electric_scale_entry)(nox_object_t*);
typedef int (*player_damage_block_wear_entry)(nox_object_t*, nox_object_t*, nox_object_t*, uint32_t, float, int);
_Static_assert(_Generic(&sub_4E2220, player_damage_electric_scale_entry: 1, default: 0), "electric armor scale ABI");
_Static_assert(_Generic(&sub_4E22A0, player_damage_block_wear_entry: 1, default: 0), "weapon block wear ABI");
_Static_assert(_Generic(&sub_4E2330, player_damage_block_wear_entry: 1, default: 0), "armor block wear ABI");

static double player_damage_guard_call(nox_object_t* target, nox_object_t* source,
	nox_object_t* weapon, int* weapon_broken, int* armor_broken) {
	player_damage_electric_scale_entry scale = sub_4E2220;
	player_damage_block_wear_entry weapon_wear = sub_4E22A0;
	player_damage_block_wear_entry armor_wear = sub_4E2330;
	double value = scale(target);
	*weapon_broken = weapon_wear(target, source, weapon, 2, 0, 0);
	*armor_broken = armor_wear(target, source, weapon, 2, 0, 0);
	return value;
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

func playerDamageGuardPublicCalls4E2220(target, source, weapon *server.Object) (float64, int, int) {
	var weaponBroken, armorBroken C.int
	scale := C.player_damage_guard_call(asObjectC(target), asObjectC(source), asObjectC(weapon), &weaponBroken, &armorBroken)
	return float64(scale), int(weaponBroken), int(armorBroken)
}
