package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type telekinesisCastHooks52D330[O comparable] struct {
	target    func() O
	class     func(O) uint32
	newObject func(string) O
	positionY func(O) float32
	positionX func(O) float32
	createAt  func(O, O, types.Pointf)
	fps       func() uint32
	apply     func(O, int32, int16, int8)
	cancel    func(int32, O)
	onSound   func(int32) sound.ID
	audio     func(sound.ID, O, int, uint32)
}

// telekinesisCast52D330 restores GAME.EXE 0052D330..0052D3BE. A missing
// target returns zero, while a non-player or failed hand allocation succeeds
// without effects. Each later operation reloads the acceptance target, except
// that Apply caches it before FPS and Audio caches it before sound lookup.
// The buff callee consumes only the low WORD duration and signed BYTE power;
// the original's full DWORD FPS multiplication wraps before that narrowing.
func telekinesisCast52D330[O comparable](id, power int32, h telekinesisCastHooks52D330[O]) int32 {
	var nilObject O
	target := h.target()
	if target == nilObject {
		return 0
	}
	if h.class(target)&uint32(object.ClassPlayer) == 0 {
		return 1
	}
	hand := h.newObject("TelekinesisHand")
	if hand == nilObject {
		return 1
	}
	target = h.target()
	y, x := h.positionY(target), h.positionX(target)
	h.createAt(hand, target, types.Ptf(x, y))
	target = h.target()
	duration := int16(20 * h.fps())
	h.apply(target, 24, duration, int8(power))
	h.cancel(24, h.target())
	h.cancel(43, h.target())
	target = h.target()
	snd := h.onSound(id)
	h.audio(snd, target, 0, 0)
	return 1
}

type TelekinesisCastRuntime52D330 struct {
	CreateAt  func(*Object, *Object, types.Pointf)
	BuffApply func(*Object, int32, int16, int8)
}

func (s *Server) CastTelekinesis52D330(id int32, _, _, _ *Object, arg *SpellAcceptArg, power int32, runtime TelekinesisCastRuntime52D330) int32 {
	return telekinesisCast52D330(id, power, telekinesisCastHooks52D330[*Object]{
		target:    func() *Object { return arg.Obj },
		class:     func(unit *Object) uint32 { return uint32(unit.ObjClass) },
		newObject: s.NewObjectByTypeID,
		positionY: func(unit *Object) float32 { return unit.PosVec.Y },
		positionX: func(unit *Object) float32 { return unit.PosVec.X },
		createAt:  runtime.CreateAt,
		fps:       s.TickRate,
		apply:     runtime.BuffApply,
		cancel: func(id int32, unit *Object) {
			s.Spells.Dur.SpellCancelDurSpell4FEB10(id, unit)
		},
		onSound: func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetOnSound() },
		audio: func(id sound.ID, unit *Object, kind int, code uint32) {
			s.Audio.EventObj(id, unit, kind, code)
		},
	})
}
