package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

type modifierResourceRuntime struct {
	damage       func(target, owner, weapon *server.Object, amount int32, kind object.DamageType)
	pointFX      func(netmsg.Op, *types.Pointf)
	audio        func(sound.ID, *server.Object)
	getMana      func(*server.Object) uint16
	getHealth    func(*server.Object) uint16
	manaSub      func(*server.Object, int32)
	manaAdd      func(*server.Object, int16)
	heal         func(*server.Object, int32)
	healthDamage func(*server.Object, int32)
}

var modifierResourceServices = modifierResourceRuntime{
	damage: func(target, owner, weapon *server.Object, amount int32, kind object.DamageType) {
		target.CallDamage(owner, weapon, int(amount), kind)
	},
	pointFX: func(op netmsg.Op, pos *types.Pointf) {
		server.NetSendPointFXNative522FF0(byte(op), pos, pointFXRuntimeForCall522FF0())
	},
	audio: func(id sound.ID, unit *server.Object) {
		GetServer().S().Audio.EventObj(id, unit, 0, 0)
	},
	getMana:   Nox_xxx_unitGetOldMana_4EEC80,
	getHealth: server.UnitGetHP4EE780,
	manaSub: func(unit *server.Object, amount int32) {
		playerManaSubCall4EEBF0(unit, amount)
	},
	manaAdd: func(unit *server.Object, amount int16) {
		playerManaAddCall4EEB80(unit, int32(amount))
	},
	heal:         unitAdjustHPCall4EE460,
	healthDamage: unitDamageClearCall4EE5E0,
}

func modifierLightning4E06F0(effect *server.ModifierEff, weapon, owner, target *server.Object, rt modifierResourceRuntime) {
	if effect == nil || target == nil {
		return
	}
	if target.Damage != nil {
		rt.damage(target, owner, weapon, server.ModifierIntegerAmount32(1, effect.AttackPreHit52.Valf), object.DamageElectric)
	}
	rt.pointFX(netmsg.MSG_FX_BLUE_SPARKS, &target.PosVec)
	rt.audio(sound.SoundWeaponEffectElectricity, target)
}

func modifierResourceTarget(effect *server.ModifierEff, owner, target *server.Object, damage *int32) bool {
	return effect != nil && owner != nil && target != nil && damage != nil && target.ObjClass.HasAny(modifierUnitClasses)
}

func modifierDrainMana4E0740(effect *server.ModifierEff, owner, target *server.Object, damage *int32, rt modifierResourceRuntime) {
	if !modifierResourceTarget(effect, owner, target, damage) {
		return
	}
	amount := server.ModifierIntegerAmount32(*damage, effect.AttackPreDmg64.Valf)
	if amount == 0 {
		amount = 1
	}
	amount = min(amount, int32(rt.getMana(target)))
	rt.manaSub(target, amount)
	rt.manaAdd(owner, int16(amount))
}

func modifierVampirism4E07C0(effect *server.ModifierEff, owner, target *server.Object, damage *int32, rt modifierResourceRuntime) {
	if !modifierResourceTarget(effect, owner, target, damage) ||
		target.ObjClass.Has(object.ClassMonster) && uint32(target.ObjSubClass)&0x40 != 0 {
		return
	}
	amount := server.ModifierIntegerAmount32(*damage, effect.AttackPreDmg64.Valf)
	if amount == 0 {
		amount = 1
	}
	rt.heal(owner, min(amount, int32(rt.getHealth(target))))
}

func modifierSympathy4E08E0(effect *server.ModifierEff, owner, target *server.Object, damage *int32, rt modifierResourceRuntime) {
	if !modifierResourceTarget(effect, owner, target, damage) {
		return
	}
	coefficient := effect.AttackPreDmg64.Valf
	amount := *damage
	amount = min(amount, int32(rt.getHealth(target)))
	rt.healthDamage(owner, server.ModifierIntegerAmount32(amount, coefficient))
}

//export nox_xxx_lightngEffect_4E06F0
func nox_xxx_lightngEffect_4E06F0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, context unsafe.Pointer) {
	modifierLightning4E06F0((*server.ModifierEff)(effect), asObjectS(weapon), asObjectS(owner), asObjectS(target), modifierResourceServices)
}

//export nox_xxx_drainMEffect_4E0740
func nox_xxx_drainMEffect_4E0740(effect unsafe.Pointer, weapon, owner, target *nox_object_t, damage *C.int32_t) {
	modifierDrainMana4E0740((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target), (*int32)(unsafe.Pointer(damage)), modifierResourceServices)
}

//export nox_xxx_vampirismEffect_4E07C0
func nox_xxx_vampirismEffect_4E07C0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, damage *C.int32_t) {
	modifierVampirism4E07C0((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target), (*int32)(unsafe.Pointer(damage)), modifierResourceServices)
}

//export nox_xxx_sympathyEffect_4E08E0
func nox_xxx_sympathyEffect_4E08E0(effect unsafe.Pointer, weapon, owner, target *nox_object_t, damage *C.int32_t) {
	modifierSympathy4E08E0((*server.ModifierEff)(effect), asObjectS(owner), asObjectS(target), (*int32)(unsafe.Pointer(damage)), modifierResourceServices)
}
