package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

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
