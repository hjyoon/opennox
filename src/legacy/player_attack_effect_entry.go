package legacy

/*
#include "GAME4_3.h"

typedef int (*pre_hit_entry)(nox_object_t*, nox_object_t*, nox_object_t*, void*);
typedef int (*item_attack_entry)(nox_object_t*, nox_object_t*, void*);
_Static_assert(_Generic(&nox_xxx_playerPreAttackEffects_538290, pre_hit_entry: 1, default: 0), "pre-hit ABI");
_Static_assert(_Generic(&nox_xxx_itemApplyAttackEffect_538840, item_attack_entry: 1, default: 0), "item attack ABI");

static int call_pre_hit_native(int indirect, nox_object_t* target, nox_object_t* owner, nox_object_t* weapon, void* attack) {
	pre_hit_entry volatile entry = nox_xxx_playerPreAttackEffects_538290;
	return indirect ? entry(target, owner, weapon, attack) : nox_xxx_playerPreAttackEffects_538290(target, owner, weapon, attack);
}

static int call_item_attack_native(int indirect, nox_object_t* weapon, nox_object_t* owner, void* attack) {
	item_attack_entry volatile entry = nox_xxx_itemApplyAttackEffect_538840;
	return indirect ? entry(weapon, owner, attack) : nox_xxx_itemApplyAttackEffect_538840(weapon, owner, attack);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerPreAttackEffects_538290
func nox_xxx_playerPreAttackEffects_538290(target, owner, weapon *nox_object_t, attack unsafe.Pointer) C.int {
	return C.int(GetServer().S().PlayerPreAttackEffects538290(asObjectS(target), asObjectS(owner), asObjectS(weapon), attack))
}

//export nox_xxx_itemApplyAttackEffect_538840
func nox_xxx_itemApplyAttackEffect_538840(weapon, owner *nox_object_t, attack unsafe.Pointer) C.int {
	return C.int(GetServer().S().ItemApplyAttackEffect538840(asObjectS(weapon), asObjectS(owner), attack))
}

func callPreHitEffectC538290(indirect bool, target, owner, weapon *server.Object, attack unsafe.Pointer) int {
	return int(C.call_pre_hit_native(C.int(bool2int(indirect)), asObjectC(target), asObjectC(owner), asObjectC(weapon), attack))
}

func callItemAttackEffectC538840(indirect bool, weapon, owner *server.Object, attack unsafe.Pointer) int {
	return int(C.call_item_attack_native(C.int(bool2int(indirect)), asObjectC(weapon), asObjectC(owner), attack))
}
