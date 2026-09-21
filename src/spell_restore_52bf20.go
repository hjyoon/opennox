package opennox

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type restoreHealthHooks52BF20[O comparable, A any] struct {
	loadTarget func(A) O
	setMaxHP   func(O)
	audio      func(sound.ID, O)
}

// restoreHealth52BF20 preserves GAME.EXE 0052BF20 without decoding
// SpellAcceptArg.Obj through the original PE32 int slot.
func restoreHealth52BF20[O comparable, A any](arg A, hooks restoreHealthHooks52BF20[O, A]) int {
	var nilObject O
	target := hooks.loadTarget(arg)
	if target == nilObject {
		return 0
	}
	hooks.setMaxHP(target)
	hooks.audio(sound.SoundRestoreHealth, target)
	return 1
}

type restoreManaHooks52BF50[O comparable, A any] struct {
	loadTarget   func(A) O
	loadClassLow func(O) uint8
	loadMaxMana  func(O) uint16
	addMana      func(O, uint16)
	audio        func(sound.ID, O)
}

// restoreMana52BF50 preserves GAME.EXE 0052BF50 while keeping the target on
// the host pointer width. Non-player targets still report success, as in the
// original callback, but receive neither mana nor audio.
func restoreMana52BF50[O comparable, A any](arg A, hooks restoreManaHooks52BF50[O, A]) int {
	var nilObject O
	target := hooks.loadTarget(arg)
	if target == nilObject {
		return 0
	}
	if hooks.loadClassLow(target)&0x04 != 0 {
		hooks.addMana(target, hooks.loadMaxMana(target))
		hooks.audio(sound.SoundRestoreMana, target)
	}
	return 1
}

func castRestoreHealth52BF20(
	_ spell.ID,
	_, _, _ *server.Object,
	arg *server.SpellAcceptArg,
	_ int,
) int {
	s := noxServer
	return restoreHealth52BF20(arg, restoreHealthHooks52BF20[*server.Object, *server.SpellAcceptArg]{
		loadTarget: func(arg *server.SpellAcceptArg) *server.Object {
			return arg.Obj
		},
		setMaxHP: legacy.Nox_xxx_unitHPsetOnMax_4EE6F0,
		audio: func(id sound.ID, target *server.Object) {
			s.Audio.EventObj(id, target, 0, 0)
		},
	})
}

func castRestoreMana52BF50(
	_ spell.ID,
	_, _, _ *server.Object,
	arg *server.SpellAcceptArg,
	_ int,
) int {
	s := noxServer
	return restoreMana52BF50(arg, restoreManaHooks52BF50[*server.Object, *server.SpellAcceptArg]{
		loadTarget: func(arg *server.SpellAcceptArg) *server.Object {
			return arg.Obj
		},
		loadClassLow: func(target *server.Object) uint8 {
			return uint8(target.ObjClass)
		},
		loadMaxMana: func(target *server.Object) uint16 {
			return target.UpdateDataPlayer().ManaMax
		},
		addMana: func(target *server.Object, amount uint16) {
			legacy.Nox_xxx_playerManaAdd_4EEB80(target, int(amount))
		},
		audio: func(id sound.ID, target *server.Object) {
			s.Audio.EventObj(id, target, 0, 0)
		},
	})
}
