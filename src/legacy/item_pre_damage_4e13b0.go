package legacy

/*
#include "GAME3_2.h"

void nullsub_39(void);
void nox_xxx_drainMEffect_4E0740(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
void nox_xxx_vampirismEffect_4E07C0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
void nox_xxx_poisonEffect_4E0850(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_sympathyEffect_4E08E0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type itemPreDamageFunctions4E13B0 struct {
	drainMana unsafe.Pointer
	vampirism unsafe.Pointer
	poison    unsafe.Pointer
	panic     unsafe.Pointer
	sympathy  unsafe.Pointer
}

func itemPreDamageFunctionsNative4E13B0() itemPreDamageFunctions4E13B0 {
	return itemPreDamageFunctions4E13B0{
		drainMana: C.nox_xxx_drainMEffect_4E0740,
		vampirism: C.nox_xxx_vampirismEffect_4E07C0,
		poison:    C.nox_xxx_poisonEffect_4E0850,
		panic:     C.nullsub_39,
		sympathy:  C.nox_xxx_sympathyEffect_4E08E0,
	}
}

func itemPreDamageCanApplyNative4E13B0(effect *server.ModifierEff) bool {
	if effect == nil || effect.AttackPreDmg64.Fnc == nil {
		return true
	}
	fnc := effect.AttackPreDmg64.Fnc
	fns := itemPreDamageFunctionsNative4E13B0()
	return fnc == fns.drainMana || fnc == fns.vampirism || fnc == fns.poison ||
		fnc == fns.panic || fnc == fns.sympathy || unsafe.Sizeof(uintptr(0)) == 4
}

// itemPreDamageApplyNative4E13B0 preserves the callback contract of GAME.EXE
// 004E13B0 without narrowing object or damage pointers to the original ABI32
// integer parameters. Unknown callbacks retain the PE32 fallback only.
func itemPreDamageApplyNative4E13B0(
	effect *server.ModifierEff,
	weapon, source, target *server.Object,
	damage *int32,
) {
	if effect == nil || effect.AttackPreDmg64.Fnc == nil || !itemPreDamageCanApplyNative4E13B0(effect) {
		return
	}
	server.CallModifierAttackEffect(effect.AttackPreDmg64.Fnc, effect, weapon, source, target, unsafe.Pointer(damage))
}
