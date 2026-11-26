package legacy

/*
#include "GAME3_2.h"
#include "common__object__modifier.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func modifierDamageMultiplierPointer4E04C0() unsafe.Pointer {
	return C.nox_xxx_effectDamageMultiplier_4E04C0
}

func modifierDamageMultiplierCallNative4E04C0(effect *server.ModifierEff, damage *float32) *float32 {
	return (*float32)(unsafe.Pointer(C.nox_xxx_effectDamageMultiplier_4E04C0(
		effect.C(), nil, nil, nil, (*C.float)(unsafe.Pointer(damage)),
	)))
}

func modifierProjectileSpeedCallNative4E09B0(effect *server.ModifierEff, projectile *server.Object) *server.Object {
	return asObjectS((*nox_object_t)(C.nox_xxx_effectProjectileSpeed_4E09B0(
		effect.C(), nil, nil, nil, asObjectC(projectile),
	)))
}

func modifierArmorMultiplierCallNative4E0370(effect *server.ModifierEff, value *float32) *float32 {
	return (*float32)(unsafe.Pointer(C.sub_4E0370(effect.C(), nil, 0, nil, 0, (*C.float)(unsafe.Pointer(value)))))
}

func modifierDurabilityMultiplierCallNative4E0380(effect *server.ModifierEff, value *float32) *float32 {
	return (*float32)(unsafe.Pointer(C.sub_4E0380(effect.C(), nil, 0, nil, 0, (*C.float)(unsafe.Pointer(value)))))
}

//export sub_4E0370
func sub_4E0370(effect unsafe.Pointer, item *nox_object_t, context C.uintptr_t, target *nox_object_t, extra C.uintptr_t, value *C.float) *C.float {
	coefficient := float32(nox_modifier_effect_getDefendFloat(effect))
	*value = C.float(server.ModifierArmorMultiplier4E0370(float32(*value), coefficient))
	return value
}

//export sub_4E0380
func sub_4E0380(effect unsafe.Pointer, item *nox_object_t, context C.uintptr_t, target *nox_object_t, extra C.uintptr_t, value *C.float) *C.float {
	coefficient := float32(nox_modifier_effect_getDefendFloat(effect))
	*value = C.float(server.ModifierDurabilityMultiplier4E0380(float32(*value), coefficient))
	return value
}

func applyModifierDefendScalar(callback unsafe.Pointer, effect *server.ModifierEff, value *float32) bool {
	switch callback {
	case C.sub_4E0370:
		sub_4E0370(effect.C(), nil, 0, nil, 0, (*C.float)(unsafe.Pointer(value)))
	case C.sub_4E0380:
		sub_4E0380(effect.C(), nil, 0, nil, 0, (*C.float)(unsafe.Pointer(value)))
	case C.nullsub_40, C.nullsub_41, C.nullsub_42:
		// Reserved defend effects leave the scalar alone.
	default:
		return false
	}
	return true
}

func applyModifierDamageMultiplier4E04C0(effect *server.ModifierEff, damage *float32) {
	if effect != nil && damage != nil {
		*damage = server.ModifierScalarMultiply32(*damage, effect.Attack40.Valf)
	}
}

//export nox_xxx_effectDamageMultiplier_4E04C0
func nox_xxx_effectDamageMultiplier_4E04C0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, damage *C.float) *C.float {
	applyModifierDamageMultiplier4E04C0((*server.ModifierEff)(effect), (*float32)(unsafe.Pointer(damage)))
	return damage
}

func applyModifierProjectileSpeed4E09B0(effect *server.ModifierEff, projectile *server.Object) {
	if effect != nil && projectile != nil {
		projectile.SpeedCur = server.ModifierScalarMultiply32(projectile.SpeedCur, effect.Attack40.Valf)
	}
}

//export nox_xxx_effectProjectileSpeed_4E09B0
func nox_xxx_effectProjectileSpeed_4E09B0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, projectile *nox_object_t) *nox_object_t {
	applyModifierProjectileSpeed4E09B0((*server.ModifierEff)(effect), asObjectS(projectile))
	return projectile
}
