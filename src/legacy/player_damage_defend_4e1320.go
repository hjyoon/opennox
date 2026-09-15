package legacy

/*
#include "defs.h"

void nullsub_40(void);
void nullsub_41(void);
void nullsub_42(void);
float* sub_4E0370(void* effect, nox_object_t* item, uintptr_t a3, nox_object_t* target, uintptr_t a5, float* value);
float* sub_4E0380(void* effect, nox_object_t* item, uintptr_t a3, nox_object_t* target, uintptr_t a5, float* value);
int nox_xxx_inversionEffect_4E03D0(int a1, int a2, int a3, int a4, int a5, int* a6);
int nox_xxx_gripEffect_4E0480(int a1, int a2, int a3, int a4, int a5, int* a6);
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type playerDamageLateDefendFunctions4E1320 struct {
	armorMultiplier      unsafe.Pointer
	durabilityMultiplier unsafe.Pointer
	resilience           unsafe.Pointer
	inversion            unsafe.Pointer
	grip                 unsafe.Pointer
	breaking             unsafe.Pointer
	punctureProne        unsafe.Pointer
}

func playerDamageLateDefendFunctionsNative4E1320() playerDamageLateDefendFunctions4E1320 {
	return playerDamageLateDefendFunctions4E1320{
		armorMultiplier:      C.sub_4E0370,
		durabilityMultiplier: C.sub_4E0380,
		resilience:           C.nullsub_40,
		inversion:            C.nox_xxx_inversionEffect_4E03D0,
		grip:                 C.nox_xxx_gripEffect_4E0480,
		breaking:             C.nullsub_41,
		punctureProne:        C.nullsub_42,
	}
}

func playerDamageCanApplyLateDefendNative4E1320(effect *server.ModifierEff) bool {
	if effect == nil || effect.Defend76.Fnc == nil {
		return true
	}
	fnc := effect.Defend76.Fnc
	fns := playerDamageLateDefendFunctionsNative4E1320()
	return fnc == fns.armorMultiplier || fnc == fns.durabilityMultiplier ||
		fnc == fns.resilience || fnc == fns.inversion || fnc == fns.grip ||
		fnc == fns.breaking || fnc == fns.punctureProne ||
		unsafe.Sizeof(uintptr(0)) == 4
}

// playerDamageApplyLateDefendNative4E1320 preserves the unusual callback
// contract of GAME.EXE 004E1320. The two multiplier effects reinterpret the
// first int32 context word as binary32, while Inversion and Grip overwrite it
// with a boolean derived from the defend-collide value at the original +96.
func playerDamageApplyLateDefendNative4E1320(
	effect *server.ModifierEff,
	item, target, weapon, source *server.Object,
	damage int32,
	typ object.DamageType,
) int32 {
	if effect == nil || effect.Defend76.Fnc == nil {
		return damage
	}
	fnc := effect.Defend76.Fnc
	fns := playerDamageLateDefendFunctionsNative4E1320()
	switch fnc {
	case fns.armorMultiplier:
		value := math.Float32frombits(uint32(damage)) * effect.Defend76.Valf
		return int32(math.Float32bits(value))
	case fns.durabilityMultiplier:
		value := math.Float32frombits(uint32(damage)) * (2.0 - effect.Defend76.Valf)
		return int32(math.Float32bits(value))
	case fns.resilience, fns.breaking, fns.punctureProne:
		return damage
	case fns.inversion:
		return int32(bool2int(uint32(effect.DefendCollide88.Val) >= 1))
	case fns.grip:
		return int32(bool2int(uint32(effect.DefendCollide88.Val) < 1))
	}
	if unsafe.Sizeof(uintptr(0)) != 4 {
		return damage
	}
	context := [2]int32{damage, int32(typ)}
	ccall.CallVoidPtr6(
		fnc,
		effect.C(), item.CObj(), target.CObj(), weapon.CObj(), source.CObj(), unsafe.Pointer(&context),
	)
	return context[0]
}
