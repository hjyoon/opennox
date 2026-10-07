package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/sound"
)

type curePoisonCastHooks52CDB0[O comparable] struct {
	target     func() O
	poison     func(O) uint8
	update     func(O, int32)
	remove     func(O)
	message    func(O, string, uint8)
	onSound    func(int32) sound.ID
	audio      func(sound.ID, O, int, uint32)
	manaCost   func(int32, int32) int32
	refundMana func(O, int16) uint16
}

// curePoisonCast52CDB0 restores GAME.EXE 0052CDB0..0052CE52. The second
// selector object is used only for an unpoisoned self-cast comparison. Poison
// strength is a zero-extended BYTE compared against signed DWORD power. The
// target is reloaded after poison treatment and priority-message delivery,
// but cached before sound lookup. The refund loads its target after the cost
// callback and passes only the signed low WORD to 004FD030.
func curePoisonCast52CDB0[O comparable](id, power int32, second O, h curePoisonCastHooks52CDB0[O]) int32 {
	var nilObject O
	target := h.target()
	if target == nilObject {
		return 0
	}
	poison := h.poison(target)
	if poison != 0 {
		message := "ExecSpel.c:PoisonClean"
		if int32(poison) > power {
			h.update(target, power)
			message = "ExecSpel.c:PoisonCure"
		} else {
			h.remove(target)
		}
		h.message(h.target(), message, 0)
	} else if target == second {
		cost := h.manaCost(id, 1)
		h.refundMana(h.target(), int16(cost))
		return 1
	}
	target = h.target()
	idSound := h.onSound(id)
	h.audio(idSound, target, 0, 0)
	return 1
}

type CurePoisonCastRuntime52CDB0 struct {
	PriorityMessage func(*Object, string, uint8)
	RefundMana      func(*Object, int16) uint16
}

// CastCurePoison52CDB0 uses native-width acceptance and object pointers with
// the existing poison, message, spell-definition and audio services. Missing
// arg still faults at the original first load; a nil entry target alone is
// an effect-free zero return. No class, flags, level or poison-immunity gate
// is introduced by this selector.
func (s *Server) CastCurePoison52CDB0(id int32, second, _, _ *Object, arg *SpellAcceptArg, power int32, runtime CurePoisonCastRuntime52CDB0) int32 {
	message := runtime.PriorityMessage
	if message == nil {
		message = func(unit *Object, id string, value uint8) { s.NetPriMsgToPlayer(unit, strman.ID(id), value) }
	}
	return curePoisonCast52CDB0(id, power, second, curePoisonCastHooks52CDB0[*Object]{
		target:     func() *Object { return arg.Obj },
		poison:     func(unit *Object) uint8 { return unit.Poison540 },
		update:     s.UpdatePoison4EE8F0,
		remove:     s.RemovePoison4EE9D0,
		message:    message,
		onSound:    func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetOnSound() },
		audio:      func(id sound.ID, unit *Object, kind int, code uint32) { s.Audio.EventObj(id, unit, kind, code) },
		manaCost:   func(id, kind int32) int32 { return int32(s.Spells.ManaCost(spell.ID(id), int(kind))) },
		refundMana: runtime.RefundMana,
	})
}
