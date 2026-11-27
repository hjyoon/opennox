package cfixture

/*
#include "item_pre_damage_4e13b0.h"
#include "item_defend_effects_4e1320.h"

static int item_pre_damage_entry(nox_object_t* target, nox_object_t* source, nox_object_t* weapon, int32_t* damage) {
	int (*volatile entry)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t*) = nox_xxx_itemApplyPreDamageEffect_4E13B0;
	return entry(target, source, weapon, damage);
}

static int item_defend_effects_entry(nox_object_t* target, nox_object_t* source, nox_object_t* weapon, int32_t* damage, int32_t damage_type) {
	int (*volatile entry)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t*, int32_t) = nox_xxx_itemApplyDefendEffect2_4E1320;
	return entry(target, source, weapon, damage, damage_type);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func ItemPreDamageC4E13B0(target, source, weapon *server.Object, damage *int32) int32 {
	return int32(C.item_pre_damage_entry(
		(*C.nox_object_t)(target.CObj()), (*C.nox_object_t)(source.CObj()), (*C.nox_object_t)(weapon.CObj()),
		(*C.int32_t)(unsafe.Pointer(damage)),
	))
}

func ItemDefendEffectsC4E1320(target, source, weapon *server.Object, damage *int32, typ object.DamageType) int32 {
	return int32(C.item_defend_effects_entry(
		(*C.nox_object_t)(target.CObj()), (*C.nox_object_t)(source.CObj()), (*C.nox_object_t)(weapon.CObj()),
		(*C.int32_t)(unsafe.Pointer(damage)), C.int32_t(typ),
	))
}
