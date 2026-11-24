package legacy

/*
#include "GAME3_2.h"
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

//export nox_modifierArmorMultiplierNative4E0370
func nox_modifierArmorMultiplierNative4E0370(effect unsafe.Pointer, value *C.float) {
	coefficient := float32(nox_modifier_effect_getDefendFloat(effect))
	*value = C.float(server.ModifierArmorMultiplier4E0370(float32(*value), coefficient))
}

//export nox_modifierDurabilityMultiplierNative4E0380
func nox_modifierDurabilityMultiplierNative4E0380(effect unsafe.Pointer, value *C.float) {
	coefficient := float32(nox_modifier_effect_getDefendFloat(effect))
	*value = C.float(server.ModifierDurabilityMultiplier4E0380(float32(*value), coefficient))
}

//export nox_modifierDamageMultiplierNative4E04C0
func nox_modifierDamageMultiplierNative4E04C0(effect unsafe.Pointer, damage *C.float) {
	if effect != nil && damage != nil {
		coefficient := (*server.ModifierEff)(effect).Attack40.Valf
		*damage = C.float(server.ModifierScalarMultiply32(float32(*damage), coefficient))
	}
}

func applyModifierProjectileSpeed4E09B0(effect *server.ModifierEff, projectile *server.Object) {
	if effect != nil && projectile != nil {
		projectile.SpeedCur = server.ModifierScalarMultiply32(projectile.SpeedCur, effect.Attack40.Valf)
	}
}

//export nox_modifierProjectileSpeedNative4E09B0
func nox_modifierProjectileSpeedNative4E09B0(effect unsafe.Pointer, projectile *nox_object_t) {
	applyModifierProjectileSpeed4E09B0((*server.ModifierEff)(effect), asObjectS(projectile))
}
