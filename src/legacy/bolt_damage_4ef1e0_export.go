package legacy

/*
#include "bolt_damage_4ef1e0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func boltDamageModifier4EF1E0(modifier *C.nox_modifier_t) *server.Modifier {
	return (*server.Modifier)(unsafe.Pointer(modifier))
}

//export nox_xxx_calcBoltDamage_4EF1E0
func nox_xxx_calcBoltDamage_4EF1E0(
	strength C.int32_t,
	modifier *C.nox_modifier_t,
) C.double {
	return C.double(GetServer().S().CalcBoltDamage4EF1E0(
		int32(strength),
		boltDamageModifier4EF1E0(modifier),
	))
}

//export nox_xxx_calcBoltDamageValues_4EF1E0
func nox_xxx_calcBoltDamageValues_4EF1E0(
	strength C.int32_t,
	typeIndex C.uint32_t,
	requiredStrength C.uint16_t,
	coefficient C.float,
	minimum C.uint16_t,
) C.double {
	return C.double(GetServer().S().CalcBoltDamageValues4EF1E0(
		int32(strength),
		server.BoltDamageModifierValues4EF1E0{
			TypeIndex:        uint32(typeIndex),
			RequiredStrength: uint16(requiredStrength),
			Coefficient:      float32(coefficient),
			Minimum:          uint16(minimum),
		},
	))
}

//export nox_xxx_boltDamageModifierType_4EF1E0
func nox_xxx_boltDamageModifierType_4EF1E0(modifier *C.nox_modifier_t) C.uint32_t {
	return C.uint32_t(boltDamageModifier4EF1E0(modifier).TypeInd)
}

//export nox_xxx_boltDamageModifierMinimum_4EF1E0
func nox_xxx_boltDamageModifierMinimum_4EF1E0(modifier *C.nox_modifier_t) C.uint16_t {
	return C.uint16_t(boltDamageModifier4EF1E0(modifier).DamageMin72)
}

//export nox_xxx_boltDamageModifierRange_4EF1E0
func nox_xxx_boltDamageModifierRange_4EF1E0(modifier *C.nox_modifier_t) C.float {
	return C.float(boltDamageModifier4EF1E0(modifier).Range68)
}
