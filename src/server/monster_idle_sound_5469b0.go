package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

type monsterIdleSoundHooks5469B0 struct {
	frame    func() uint32
	tickRate func() uint32
	random   func(min, max int) int
	play     func(id sound.ID, unit *Object)
}

func monsterIdleSound5469B0(unit *Object, hooks monsterIdleSoundHooks5469B0) {
	if unit == nil || !unit.ObjClass.Has(object.ClassMonster) || unit.UpdateData == nil {
		return
	}
	unit.MonsterIdleSoundNative5469B0(hooks.frame, hooks.tickRate, hooks.random,
		func(id uint32, unit *Object, _, _ uint32) { hooks.play(sound.ID(id), unit) })
}

// MonsterIdleSound5469B0 keeps the original idle/guard sound schedule while
// reading MonsterUpdateData and the sound-set pointer at native width.
func (s *Server) MonsterIdleSound5469B0(unit *Object) {
	monsterIdleSound5469B0(unit, monsterIdleSoundHooks5469B0{
		frame:    s.Frame,
		tickRate: s.TickRate,
		random:   s.Rand.Logic.IntClamp,
		play: func(id sound.ID, unit *Object) {
			s.Audio.EventObj(id, unit, 0, 0)
		},
	})
}
