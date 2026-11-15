package legacy

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

type resourceSpellDeps52BF20 struct {
	setMaximumHP func(*server.Object)
	addMana      func(*server.Object, int32)
	hp           func(*server.Object) uint16
	maximumHP    func(*server.Object) uint16
	cost         func(int32, int32) int32
	recharge     func(*server.Object, int32)
	balance      func(string) float64
	multiplier   func(byte) float32
	convert      func(float32) int32
	adjustHP     func(*server.Object, int32)
	sound        func(int32, int32) int32
	audio        func(int32, *server.Object)
}

var resourceSpellDepsForCall52BF20 = resourceSpellDefaultDeps52BF20

func resourceSpellDefaultDeps52BF20() resourceSpellDeps52BF20 {
	return resourceSpellDeps52BF20{
		setMaximumHP: Nox_xxx_unitHPsetOnMax_4EE6F0,
		addMana: func(unit *server.Object, amount int32) {
			// This producer zero-extends the original maximum-mana word into
			// the complete DWORD slot. Keep the historical short API separate.
			Nox_xxx_playerManaAddFull4EEB80(unit, amount)
		},
		hp:        Nox_xxx_unitGetHP_4EE780,
		maximumHP: Nox_xxx_unitGetMaxHP_4EE7A0,
		cost:      nox_xxx_spellManaCost_4249A0,
		recharge: func(unit *server.Object, amount int32) {
			// The recharge callee consumes the signed low word of the slot.
			playerManaRechargeCall4FD030(unit, int16(amount))
		},
		balance: func(key string) float64 { return GetServer().S().Balance.Float(key) },
		multiplier: func(class byte) float32 {
			stats := &GetServer().S().Players.Mult
			switch class {
			case 0:
				return stats.Warrior.Health
			case 2:
				return stats.Conjurer.Health
			case 1:
				return stats.Wizard.Health
			default:
				panic("unexpected resource spell class")
			}
		},
		convert:  resourceSpellFloatToInt52DD50,
		adjustHP: unitAdjustHPCall4EE460,
		sound:    nox_xxx_spellGetAud44_424800,
		audio: func(id int32, unit *server.Object) {
			GetServer().S().Audio.EventObj(sound.ID(id), unit, 0, 0)
		},
	}
}

func resourceRestoreHealth52BF20(arg *server.SpellAcceptArg, deps resourceSpellDeps52BF20) int32 {
	target := arg.Obj
	if target == nil {
		return 0
	}
	deps.setMaximumHP(target)
	deps.audio(754, arg.Obj)
	return 1
}

func resourceRestoreMana52BF50(arg *server.SpellAcceptArg, deps resourceSpellDeps52BF20) int32 {
	target := arg.Obj
	if target == nil {
		return 0
	}
	if target.Class().Has(object.ClassPlayer) {
		amount := uint16(target.UpdateDataPlayer().ManaMax)
		deps.addMana(target, int32(amount))
		deps.audio(755, arg.Obj)
	}
	return 1
}

// The original startup selects x87 ToZero. Both binary32 stores in Lesser
// Heal use that mode; a host float32 cast otherwise rounds away in some cases.
func resourceSpellFloat32ToZero52DD50(value float64) float32 {
	stored := float32(value)
	if math.Abs(float64(stored)) > math.Abs(value) {
		return math.Nextafter32(stored, 0)
	}
	return stored
}

func resourceLesserHeal52DD50(spellID int32, second, caster *server.Object, arg *server.SpellAcceptArg, deps resourceSpellDeps52BF20) int32 {
	target := arg.Obj
	if target == nil {
		return 0
	}
	currentHP := deps.hp(arg.Obj)
	maximumHP := deps.maximumHP(target)
	if currentHP == maximumHP && second == arg.Obj {
		cost := deps.cost(spellID, 1)
		deps.recharge(caster, cost)
		return 1
	}
	amount := resourceSpellFloat32ToZero52DD50(deps.balance("LesserHealAmount"))
	if caster != nil && caster.Class().Has(object.ClassPlayer) {
		player := caster.UpdateDataPlayer().Player
		// The original requires this owner. Nil-safe Player accessors must
		// not silently reinterpret missing data as the Warrior class.
		if player == nil {
			panic("nil Player in resource spell caster")
		}
		class := byte(player.Info().PlayerClass())
		if class == 0 || class == 2 || class == 1 {
			factor := deps.multiplier(class)
			amount = resourceSpellFloat32ToZero52DD50(float64(factor) * float64(amount))
		}
	}
	delta := deps.convert(amount)
	deps.adjustHP(arg.Obj, delta)
	// Sound lookup can change the live target. The original saves the audio
	// object after HP adjustment and before that lookup.
	audioTarget := arg.Obj
	id := deps.sound(spellID, 1)
	deps.audio(id, audioTarget)
	return 1
}
