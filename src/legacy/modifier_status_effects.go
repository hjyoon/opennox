package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

const modifierUnitClasses = object.ClassMonster | object.ClassPlayer

func reportModifierStatus4E04D0(player byte, status int32) {
	GetServer().S().NetInformTextMsg(ntype.PlayerInd(player), 13, int(status))
}

func applyModifierStatusCast4E04D0(
	effect *server.ModifierEff,
	owner, target *server.Object,
	status int32,
	cast func(*server.Object, *server.SpellAcceptArg, int32) int32,
	inform func(byte, int32),
) {
	if !target.ObjClass.HasAny(modifierUnitClasses) {
		return
	}
	arg := server.SpellAcceptArg{Obj: target, Pos: target.PosVec}
	power := int32(nox_modifier_effect_getPreHitInt(effect.C()))
	cast(owner, &arg, power)
	if target.ObjClass.Has(object.ClassPlayer) {
		update := (*server.PlayerUpdateData)(target.UpdateData)
		inform(update.Player.PlayerInd, status)
	}
}

//export nox_xxx_stunEffect_4E04D0
func nox_xxx_stunEffect_4E04D0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, context unsafe.Pointer) {
	applyModifierStatusCast4E04D0((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target),
		0, castStunCall52C2C0, reportModifierStatus4E04D0)
}

//export nox_xxx_confuseEffect_4E0670
func nox_xxx_confuseEffect_4E0670(effect unsafe.Pointer, weapon, owner, target *nox_object_t, context unsafe.Pointer) {
	applyModifierStatusCast4E04D0((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target),
		1, castConfuseCall52C1E0, reportModifierStatus4E04D0)
}

//export nox_xxx_recoilEffect_4E0640
func nox_xxx_recoilEffect_4E0640(effect unsafe.Pointer, weapon, owner, target *nox_object_t, context unsafe.Pointer) {
	if effect != nil && weapon != nil && target != nil {
		origin := &asObjectS(weapon).PosVec
		force := (*server.ModifierEff)(effect).AttackPreHit52.Valf
		applyForceExport52DF80(asObjectS(target), unsafe.Pointer(origin), force)
	}
}

func modifierPoisonFacing4E0850(target *types.Pointf, direction int16, source *types.Pointf) int {
	return int(twoPointsAndDirection4E6E50(*target, int32(direction), *source))
}

func applyModifierPoison4E0850(
	effect *server.ModifierEff,
	owner, target *server.Object,
	facing func(*types.Pointf, int16, *types.Pointf) int,
	activate func(*server.Object, int32, int32) int32,
	inform func(byte, int32),
) {
	if effect == nil || target == nil {
		return
	}
	if target.ObjClass.Has(object.ClassPlayer) && target.UpdateData != nil && owner != nil {
		update := (*server.PlayerUpdateData)(target.UpdateData)
		if update.State == 16 && facing(&target.PosVec, int16(target.Direction1), &owner.PosVec)&1 != 0 {
			return
		}
	}
	if !target.ObjClass.HasAny(modifierUnitClasses) || activate(target, 1, effect.AttackPreDmg64.Val) == 0 {
		return
	}
	if target.ObjClass.Has(object.ClassPlayer) && target.UpdateData != nil {
		update := (*server.PlayerUpdateData)(target.UpdateData)
		if update.Player != nil {
			inform(update.Player.PlayerInd, 2)
		}
	}
}

func modifierPoison4E0850(effect *server.ModifierEff, owner, target *server.Object) {
	applyModifierPoison4E0850(effect, owner, target,
		modifierPoisonFacing4E0850, Nox_xxx_activatePoison_4EE7E0, reportModifierStatus4E04D0)
}

//export nox_xxx_poisonEffect_4E0850
func nox_xxx_poisonEffect_4E0850(effect unsafe.Pointer, weapon, owner, target *nox_object_t, context unsafe.Pointer) {
	modifierPoison4E0850((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target))
}
