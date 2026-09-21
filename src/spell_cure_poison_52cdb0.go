package opennox

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

const (
	curePoisonReducedMessage52CDB0 = "ExecSpel.c:PoisonCure"
	curePoisonClearedMessage52CDB0 = "ExecSpel.c:PoisonClean"
)

type curePoisonHooks52CDB0[O comparable] struct {
	loadPoison func(O) uint8
	update     func(O, int32)
	remove     func(O)
	message    func(O, string)
	audio      func(spell.ID, O)
	manaCost   func(spell.ID, int) int
	refundMana func(O, int)
}

// curePoison52CDB0 preserves GAME.EXE 0052CDB0 without narrowing the caster,
// target, or SpellAcceptArg object pointer to a PE32 int. A strong enough cast
// clears poison, a weaker cast reduces its strength, and a no-op self cast
// refunds the level-one mana cost exactly as the original does.
func curePoison52CDB0[O comparable](
	spellID spell.ID,
	caster, target O,
	level int32,
	hooks curePoisonHooks52CDB0[O],
) int {
	var nilObject O
	if target == nilObject {
		return 0
	}
	if poison := hooks.loadPoison(target); poison != 0 {
		if int32(poison) > level {
			hooks.update(target, level)
			hooks.message(target, curePoisonReducedMessage52CDB0)
		} else {
			hooks.remove(target)
			hooks.message(target, curePoisonClearedMessage52CDB0)
		}
		hooks.audio(spellID, target)
		return 1
	}
	if target != caster {
		hooks.audio(spellID, target)
		return 1
	}
	hooks.refundMana(target, hooks.manaCost(spellID, 1))
	return 1
}

func castCurePoison(
	spellID spell.ID,
	caster, _, _ *server.Object,
	arg *server.SpellAcceptArg,
	level int,
) int {
	s := noxServer
	return curePoison52CDB0(spellID, caster, arg.Obj, int32(level), curePoisonHooks52CDB0[*server.Object]{
		loadPoison: func(target *server.Object) uint8 {
			return target.Poison540
		},
		update: func(target *server.Object, amount int32) {
			s.S().UpdatePoison4EE8F0(target, amount)
		},
		remove: func(target *server.Object) {
			s.S().RemovePoison4EE9D0(target)
		},
		message: func(target *server.Object, message string) {
			s.NetPriMsgToPlayer(target, strman.ID(message), 0)
		},
		audio: func(id spell.ID, target *server.Object) {
			aud := s.Spells.DefByInd(id).GetOnSound()
			s.Audio.EventObj(aud, target, 0, 0)
		},
		manaCost: func(id spell.ID, level int) int {
			return s.Spells.ManaCost(id, level)
		},
		refundMana: func(target *server.Object, amount int) {
			sub_4FD030(target, amount)
		},
	})
}
